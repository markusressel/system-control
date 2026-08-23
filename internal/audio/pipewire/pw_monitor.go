package pipewire

import (
	"encoding/json"
	"log"
	"os/exec"
	"sync"
	"time"
)

var (
	liveState       *GraphState
	liveObjects     map[int]GraphObject
	liveMu          sync.RWMutex
	monitorStarted  bool
	updateCallbacks []func()

	isStateDirty bool
	debounceChan chan struct{}
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
	liveMu.RUnlock()

	return pwDumpDirect()
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

func runMonitorLoop() {
	cmd := exec.Command("pw-dump", "-m")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to get stdout pipe for pw-dump -m: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start pw-dump -m: %v", err)
		return
	}

	decoder := json.NewDecoder(stdout)

	for {
		var chunk []RawGraphObject
		if err := decoder.Decode(&chunk); err != nil {
			log.Printf("pw-dump -m decoder error: %v", err)
			break
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
				mergedMetadata = obj.Metadata
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

		select {
		case debounceChan <- struct{}{}:
		default:
		}
	}
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
			newState.Nodes = append(newState.Nodes, InterfaceNode{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceNodeInfo),
			})
		case TypeFactory:
			newState.Factories = append(newState.Factories, InterfaceFactory{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceFactoryInfo),
			})
		case TypeModule:
			newState.Modules = append(newState.Modules, InterfaceModule{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceModuleInfo),
			})
		case TypeCore:
			newState.Cores = append(newState.Cores, InterfaceCore{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceCoreInfo),
			})
		case TypeClient:
			newState.Clients = append(newState.Clients, InterfaceClient{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceClientInfo),
			})
		case TypeLink:
			newState.Links = append(newState.Links, InterfaceLink{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceLinkInfo),
			})
		case TypePort:
			newState.Ports = append(newState.Ports, InterfacePort{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfacePortInfo),
			})
		case TypeDevice:
			newState.Devices = append(newState.Devices, InterfaceDevice{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceDeviceInfo),
			})
		}
	}
	liveState = newState
}
