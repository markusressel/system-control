package pipewire

import (
	"encoding/json"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	liveState       *GraphState
	liveObjects     map[int]GraphObject
	liveMu          sync.RWMutex
	monitorStarted  bool
	updateCallbacks []func()

	isStateDirty bool
	// metadataStale is set when a metadata update cannot be merged incrementally
	metadataStale bool
	debounceChan  chan struct{}
)

// RegisterUpdateCallback allows the daemon to subscribe to PipeWire graph changes
func RegisterUpdateCallback(cb func()) {
	liveMu.Lock()
	defer liveMu.Unlock()
	updateCallbacks = append(updateCallbacks, cb)
}

// RawGraphObject captures the raw JSON of an object to merge incremental diffs properly.
type RawGraphObject struct {
	Id          int                      `json:"id"`
	Type        string                   `json:"type"`
	Version     int                      `json:"version"`
	Permissions []string                 `json:"permissions"`
	Props       map[string]interface{}   `json:"props,omitempty"`
	Metadata    []map[string]interface{} `json:"metadata,omitempty"`
	Info        json.RawMessage          `json:"info,omitempty"`
}

func StartMonitor() {
	liveMu.Lock()
	if !monitorStarted {
		monitorStarted = true
		liveObjects = make(map[int]GraphObject)
		debounceChan = make(chan struct{}, 1)
		go runMonitorLoop()
		go debounceLoop()
	}
	liveMu.Unlock()
}

func PwDump() GraphState {
	liveMu.RLock()
	dirty := isStateDirty
	liveMu.RUnlock()

	if dirty {
		liveMu.Lock()
		if isStateDirty {
			rebuildLiveState()
			isStateDirty = false
		}
		liveMu.Unlock()
	}

	liveMu.RLock()
	if monitorStarted && liveState != nil {
		defer liveMu.RUnlock()
		return *liveState
	}
	monitored := monitorStarted
	liveMu.RUnlock()

	state, err := pwDumpDirect()
	if err != nil {
		if monitored {
			// a long running daemon must survive PipeWire being unavailable
			log.Printf("pw-dump failed: %v", err)
			return GraphState{}
		}
		log.Fatal(err)
	}
	return state
}

func debounceLoop() {
	coolDown := time.Millisecond * 50
	for range debounceChan {
		// Fire callbacks immediately for 0ms delay
		var callbacks []func()
		liveMu.RLock()
		callbacks = append(callbacks, updateCallbacks...)
		liveMu.RUnlock()

		for _, cb := range callbacks {
			cb()
		}

		// Then sleep to throttle subsequent spam (max 20 FPS)
		time.Sleep(coolDown)
	}
}

// runMonitorLoop keeps a `pw-dump -m` process running. When it dies (e.g. PipeWire restarts)
// the live state is dropped, so PwDump falls back to direct queries, and the monitor is restarted.
func runMonitorLoop() {
	backoff := time.Second
	for {
		started := time.Now()
		monitorOnce()

		liveMu.Lock()
		liveObjects = make(map[int]GraphObject)
		liveState = nil
		isStateDirty = false
		metadataStale = false
		liveMu.Unlock()
		notifyChanged()

		if time.Since(started) > time.Minute {
			backoff = time.Second
		}
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func notifyChanged() {
	select {
	case debounceChan <- struct{}{}:
	default:
	}
}

// monitorOnce runs a single `pw-dump -m` process until it ends or its output cannot be decoded.
func monitorOnce() {
	cmd := exec.Command("pw-dump", "-m")
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to get stdout pipe for pw-dump -m: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start pw-dump -m: %v", err)
		return
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	decoder := json.NewDecoder(stdout)

	for {
		var chunk []RawGraphObject
		if err := decoder.Decode(&chunk); err != nil {
			log.Printf("pw-dump -m decoder error: %v", err)
			return
		}

		liveMu.Lock()
		for _, obj := range chunk {
			if string(obj.Info) == "null" || (obj.Info == nil && obj.Type == "") {
				// Object removed
				delete(liveObjects, obj.Id)
				continue
			}

			// If type is missing (incremental update), inherit from existing
			existing, exists := liveObjects[obj.Id]

			objType := obj.Type
			if !exists {
				// New object
			} else {
				// Inherit missing fields
				if objType == "" {
					objType = existing.Type
				}
				if obj.Version == 0 {
					obj.Version = existing.Version
				}
				if obj.Permissions == nil {
					obj.Permissions = existing.Permissions
				}
			}

			// Parse Info
			var parsedInfo GraphObjectInfo
			if obj.Info != nil {
				parsedInfo = parseInfoMerge(objType, obj.Info, exists, existing.Info)
			} else if exists {
				parsedInfo = existing.Info
			}

			var mergedProps map[string]interface{}
			if obj.Props != nil {
				mergedProps = obj.Props // Note: simple overwrite, usually fine for top-level Props
			} else if exists {
				mergedProps = existing.Props
			}

			var mergedMetadata []map[string]interface{}
			if obj.Metadata != nil {
				if exists {
					var stale bool
					mergedMetadata, stale = mergeMetadata(existing.Metadata, obj.Metadata)
					metadataStale = metadataStale || stale
				} else {
					mergedMetadata = obj.Metadata
				}
			} else if exists {
				mergedMetadata = existing.Metadata
			}

			goObj := GraphObject{
				CommonData: CommonData{
					Id:          obj.Id,
					Type:        objType,
					Version:     obj.Version,
					Permissions: obj.Permissions,
					Props:       mergedProps,
					Metadata:    mergedMetadata,
				},
				Info: parsedInfo,
			}

			liveObjects[obj.Id] = goObj
		}

		isStateDirty = true
		liveMu.Unlock()

		notifyChanged()
	}
}

// mergeMetadata applies an incremental metadata update. pw-dump only reports the changed
// entries, and reports removals as empty or null entries without the removed key. Such updates
// can't be merged, stale is true then and the metadata has to be queried again.
func mergeMetadata(existing []map[string]interface{}, update []map[string]interface{}) (merged []map[string]interface{}, stale bool) {
	if len(update) == 0 {
		return existing, true
	}

	type entryKey struct {
		subject interface{}
		key     interface{}
	}
	index := map[entryKey]int{}
	for i, entry := range existing {
		index[entryKey{entry["subject"], entry["key"]}] = i
	}

	merged = append(merged, existing...)
	for _, entry := range update {
		if entry["value"] == nil {
			stale = true
		}
		k := entryKey{entry["subject"], entry["key"]}
		if i, ok := index[k]; ok {
			merged[i] = entry
		} else {
			index[k] = len(merged)
			merged = append(merged, entry)
		}
	}
	return merged, stale
}

func parseInfoMerge(objType string, rawInfo []byte, exists bool, existing GraphObjectInfo) GraphObjectInfo {
	switch objType {
	case TypeNode:
		var info InterfaceNodeInfo
		if exists && existing != nil {
			info = existing.(InterfaceNodeInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeFactory:
		var info InterfaceFactoryInfo
		if exists && existing != nil {
			info = existing.(InterfaceFactoryInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeModule:
		var info InterfaceModuleInfo
		if exists && existing != nil {
			info = existing.(InterfaceModuleInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeCore:
		var info InterfaceCoreInfo
		if exists && existing != nil {
			info = existing.(InterfaceCoreInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeClient:
		var info InterfaceClientInfo
		if exists && existing != nil {
			info = existing.(InterfaceClientInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeLink:
		var info InterfaceLinkInfo
		if exists && existing != nil {
			info = existing.(InterfaceLinkInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypePort:
		var info InterfacePortInfo
		if exists && existing != nil {
			info = existing.(InterfacePortInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeDevice:
		var info InterfaceDeviceInfo
		if exists && existing != nil {
			info = existing.(InterfaceDeviceInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeProfiler:
		var info InterfaceProfilerInfo
		if exists && existing != nil {
			info = existing.(InterfaceProfilerInfo)
		} else {
			info = InterfaceProfilerInfo(&map[string]interface{}{})
		}
		json.Unmarshal(rawInfo, &info)
		return info
	case TypeMetadata:
		var info InterfaceMetadataInfo
		if exists && existing != nil {
			info = existing.(InterfaceMetadataInfo)
		}
		json.Unmarshal(rawInfo, &info)
		return info
	}
	return nil
}

// rebuildLiveState reconstitutes a typed GraphState from the raw objects map
// Must be called with liveMu Write Lock held.
func rebuildLiveState() {
	newState := &GraphState{}
	for _, object := range liveObjects {
		switch object.CommonData.Type {
		case TypeNode:
			if info, ok := object.Info.(InterfaceNodeInfo); ok {
				newState.Nodes = append(newState.Nodes, InterfaceNode{CommonData: object.CommonData, Info: info})
			}
		case TypeFactory:
			if info, ok := object.Info.(InterfaceFactoryInfo); ok {
				newState.Factories = append(newState.Factories, InterfaceFactory{CommonData: object.CommonData, Info: info})
			}
		case TypeModule:
			if info, ok := object.Info.(InterfaceModuleInfo); ok {
				newState.Modules = append(newState.Modules, InterfaceModule{CommonData: object.CommonData, Info: info})
			}
		case TypeCore:
			if info, ok := object.Info.(InterfaceCoreInfo); ok {
				newState.Cores = append(newState.Cores, InterfaceCore{CommonData: object.CommonData, Info: info})
			}
		case TypeClient:
			if info, ok := object.Info.(InterfaceClientInfo); ok {
				newState.Clients = append(newState.Clients, InterfaceClient{CommonData: object.CommonData, Info: info})
			}
		case TypeLink:
			if info, ok := object.Info.(InterfaceLinkInfo); ok {
				newState.Links = append(newState.Links, InterfaceLink{CommonData: object.CommonData, Info: info})
			}
		case TypePort:
			if info, ok := object.Info.(InterfacePortInfo); ok {
				newState.Ports = append(newState.Ports, InterfacePort{CommonData: object.CommonData, Info: info})
			}
		case TypeDevice:
			if info, ok := object.Info.(InterfaceDeviceInfo); ok {
				newState.Devices = append(newState.Devices, InterfaceDevice{CommonData: object.CommonData, Info: info})
			}
		case TypeMetadata:
			// metadata (e.g. the default sink) lives in CommonData.Metadata, there is no Info
			info, _ := object.Info.(InterfaceMetadataInfo)
			newState.Metadatas = append(newState.Metadatas, InterfaceMetadata{CommonData: object.CommonData, Info: info})
		}
	}

	if metadataStale {
		if fresh, err := pwDumpDirect(); err == nil {
			newState.Metadatas = fresh.Metadatas
			metadataStale = false
		}
	}
	liveState = newState
}
