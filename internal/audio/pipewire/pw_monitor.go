package pipewire

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/markusressel/system-control/internal/supervise"
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
	liveMu.RUnlock()

	state, err := pwDumpDirect()
	if err != nil {
		// callers see an empty graph, a long running daemon must survive PipeWire being unavailable
		log.Printf("pw-dump failed: %v", err)
		return GraphState{}
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

// monitor keeps a `pw-dump -m` process running.
var monitor = &supervise.Loop{Name: "pipewire: pw-dump -m"}

// MonitorStatus describes the health of the `pw-dump -m` monitor for `daemon status`.
func MonitorStatus() []string {
	return monitor.Status()
}

// runMonitorLoop keeps a `pw-dump -m` process running. When it dies (e.g. PipeWire restarts)
// the live state is dropped, so PwDump falls back to direct queries, and the monitor is restarted.
func runMonitorLoop() {
	monitor.Run(context.Background(), func(ctx context.Context, ready func()) error {
		err := monitorOnce(ctx, ready)
		dropLiveState()
		return err
	}, nil)
}

// dropLiveState forgets the objects of a monitor that ended.
func dropLiveState() {
	liveMu.Lock()
	liveObjects = make(map[int]GraphObject)
	liveState = nil
	isStateDirty = false
	metadataStale = false
	liveMu.Unlock()
	notifyChanged()
}

func notifyChanged() {
	select {
	case debounceChan <- struct{}{}:
	default:
	}
}

// monitorOnce runs a single `pw-dump -m` process until it ends or its output cannot be decoded.
func monitorOnce(ctx context.Context, ready func()) error {
	cmd := exec.CommandContext(ctx, "pw-dump", "-m")
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()
	ready()

	decoder := json.NewDecoder(stdout)
	for {
		var chunk []RawGraphObject
		if err := decoder.Decode(&chunk); err != nil {
			return fmt.Errorf("decoding its output: %w", err)
		}
		applyChunk(chunk)
		notifyChanged()
	}
}

// applyChunk applies one batch of changes of `pw-dump -m` to the live objects.
func applyChunk(chunk []RawGraphObject) {
	liveMu.Lock()
	defer liveMu.Unlock()
	for _, obj := range chunk {
		if isRemoval(obj) {
			delete(liveObjects, obj.Id)
			continue
		}
		existing, exists := liveObjects[obj.Id]
		merged, stale := mergeObject(obj, existing, exists)
		liveObjects[obj.Id] = merged
		metadataStale = metadataStale || stale
	}
	isStateDirty = true
}

// isRemoval tells whether pw-dump reports that the object went away.
func isRemoval(obj RawGraphObject) bool {
	return string(obj.Info) == "null" || (obj.Info == nil && obj.Type == "")
}

// mergeObject applies an incremental update of pw-dump to the object it knows already (if it
// exists): fields the update leaves out keep their values. metadataStale is true if a
// metadata update could not be merged, see mergeMetadata.
func mergeObject(obj RawGraphObject, existing GraphObject, exists bool) (merged GraphObject, metadataStale bool) {
	if !exists {
		existing = GraphObject{}
	}
	objType := cmp.Or(obj.Type, existing.Type)

	info := existing.Info
	if obj.Info != nil {
		info = parseInfoMerge(objType, obj.Info, exists, existing.Info)
	}

	// top level props are replaced as a whole
	props := existing.Props
	if obj.Props != nil {
		props = obj.Props
	}

	metadata := existing.Metadata
	if obj.Metadata != nil {
		metadata = obj.Metadata
		if exists {
			metadata, metadataStale = mergeMetadata(existing.Metadata, obj.Metadata)
		}
	}

	permissions := obj.Permissions
	if permissions == nil {
		permissions = existing.Permissions
	}

	return GraphObject{
		CommonData: CommonData{
			Id:          obj.Id,
			Type:        objType,
			Version:     cmp.Or(obj.Version, existing.Version),
			Permissions: permissions,
			Props:       props,
			Metadata:    metadata,
		},
		Info: info,
	}, metadataStale
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
	if !exists {
		existing = nil
	}
	switch objType {
	case TypeNode:
		return mergeInfo[InterfaceNodeInfo](rawInfo, existing)
	case TypeFactory:
		return mergeInfo[InterfaceFactoryInfo](rawInfo, existing)
	case TypeModule:
		return mergeInfo[InterfaceModuleInfo](rawInfo, existing)
	case TypeCore:
		return mergeInfo[InterfaceCoreInfo](rawInfo, existing)
	case TypeClient:
		return mergeInfo[InterfaceClientInfo](rawInfo, existing)
	case TypeLink:
		return mergeInfo[InterfaceLinkInfo](rawInfo, existing)
	case TypePort:
		return mergeInfo[InterfacePortInfo](rawInfo, existing)
	case TypeDevice:
		return mergeInfo[InterfaceDeviceInfo](rawInfo, existing)
	case TypeProfiler:
		info := mergeInfo[InterfaceProfilerInfo](rawInfo, existing)
		if info == nil {
			info = InterfaceProfilerInfo(&map[string]interface{}{})
		}
		return info
	case TypeMetadata:
		return mergeInfo[InterfaceMetadataInfo](rawInfo, existing)
	}
	return nil
}

// mergeInfo applies an incremental info update to a copy of the existing info. The existing
// info must not be changed: its maps and slices are shared with published snapshots (PwDump),
// which are read without liveMu, and json.Unmarshal writes into existing maps and slices in
// place.
func mergeInfo[T GraphObjectInfo](rawInfo []byte, existing GraphObjectInfo) T {
	var info T
	if previous, ok := existing.(T); ok {
		if encoded, err := json.Marshal(previous); err == nil {
			_ = json.Unmarshal(encoded, &info)
		}
	}
	_ = json.Unmarshal(rawInfo, &info)
	return info
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
