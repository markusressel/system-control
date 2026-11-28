package pipewire

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/markusressel/system-control/internal/supervise"
)

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

// liveGraph is the PipeWire graph kept up to date by a `pw-dump -m` monitor, so that queries
// don't have to run pw-dump.
type liveGraph struct {
	// monitorCommand starts the monitor, queryMetadata reads the metadata again when an update
	// can't be merged. Variables for tests.
	monitorCommand func(ctx context.Context) *exec.Cmd
	queryMetadata  func() ([]InterfaceMetadata, error)

	monitor *supervise.Loop
	changed chan struct{}

	// mu guards the fields below
	mu        sync.RWMutex
	started   bool
	callbacks []func()
	objects   map[int]GraphObject
	// state is built from objects on demand (dirty), nil while there is no monitor
	state *GraphState
	dirty bool
	// metadataStale is set when a metadata update cannot be merged incrementally
	metadataStale bool
}

func newLiveGraph() *liveGraph {
	return &liveGraph{
		monitorCommand: func(ctx context.Context) *exec.Cmd { return exec.CommandContext(ctx, "pw-dump", "-m") },
		queryMetadata: func() ([]InterfaceMetadata, error) {
			state, err := pwDumpDirect()
			return state.Metadatas, err
		},
		monitor: &supervise.Loop{Name: "pipewire: pw-dump -m"},
		changed: make(chan struct{}, 1),
		objects: map[int]GraphObject{},
	}
}

// live is the graph of this process, monitored only in the daemon (StartMonitor).
var live = newLiveGraph()

// RegisterUpdateCallback allows the daemon to subscribe to PipeWire graph changes
func RegisterUpdateCallback(cb func()) {
	live.mu.Lock()
	defer live.mu.Unlock()
	live.callbacks = append(live.callbacks, cb)
}

// StartMonitor keeps the graph up to date with a `pw-dump -m` process, so that PwDump does not
// have to query PipeWire.
func StartMonitor() {
	live.start()
}

// MonitorStatus describes the health of the `pw-dump -m` monitor for `daemon status`.
func MonitorStatus() []string {
	return live.monitor.Status()
}

// PwDump returns the current graph: the monitored one, or a fresh query without a monitor.
func PwDump() GraphState {
	if state, ok := live.snapshot(); ok {
		return state
	}
	state, err := pwDumpDirect()
	if err != nil {
		// callers see an empty graph, a long running daemon must survive PipeWire being unavailable
		log.Printf("pw-dump failed: %v", err)
		return GraphState{}
	}
	return state
}

func (g *liveGraph) start() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.started {
		return
	}
	g.started = true
	go g.monitor.Run(context.Background(), func(ctx context.Context, ready func()) error {
		err := g.monitorOnce(ctx, ready)
		g.drop()
		return err
	}, nil)
	go g.notifyLoop()
}

// snapshot returns the monitored graph, ok is false while there is none.
func (g *liveGraph) snapshot() (GraphState, bool) {
	g.mu.RLock()
	dirty := g.dirty
	g.mu.RUnlock()

	if dirty {
		g.mu.Lock()
		if g.dirty {
			g.rebuild()
			g.dirty = false
		}
		g.mu.Unlock()
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.started || g.state == nil {
		return GraphState{}, false
	}
	return *g.state, true
}

// notifyLoop calls the callbacks on changes: right away, but at most 20 times a second.
func (g *liveGraph) notifyLoop() {
	const coolDown = 50 * time.Millisecond
	for range g.changed {
		g.mu.RLock()
		callbacks := slices.Clone(g.callbacks)
		g.mu.RUnlock()

		for _, cb := range callbacks {
			cb()
		}
		time.Sleep(coolDown)
	}
}

func (g *liveGraph) notifyChanged() {
	select {
	case g.changed <- struct{}{}:
	default:
	}
}

// drop forgets the objects of a monitor that ended, so PwDump falls back to direct queries.
func (g *liveGraph) drop() {
	g.mu.Lock()
	g.objects = map[int]GraphObject{}
	g.state = nil
	g.dirty = false
	g.metadataStale = false
	g.mu.Unlock()
	g.notifyChanged()
}

// monitorOnce runs a single `pw-dump -m` process until it ends or its output cannot be decoded.
func (g *liveGraph) monitorOnce(ctx context.Context, ready func()) error {
	cmd := g.monitorCommand(ctx)
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
	return g.consume(stdout)
}

// consume applies the output of `pw-dump -m` until it ends or can't be decoded.
func (g *liveGraph) consume(r io.Reader) error {
	decoder := json.NewDecoder(r)
	for {
		var chunk []RawGraphObject
		if err := decoder.Decode(&chunk); err != nil {
			return fmt.Errorf("decoding its output: %w", err)
		}
		g.applyChunk(chunk)
		g.notifyChanged()
	}
}

// applyChunk applies one batch of changes of `pw-dump -m` to the live objects.
func (g *liveGraph) applyChunk(chunk []RawGraphObject) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, obj := range chunk {
		if isRemoval(obj) {
			delete(g.objects, obj.Id)
			continue
		}
		existing, exists := g.objects[obj.Id]
		merged, stale := mergeObject(obj, existing, exists)
		g.objects[obj.Id] = merged
		g.metadataStale = g.metadataStale || stale
	}
	g.dirty = true
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

// rebuild reconstitutes a typed GraphState from the raw objects. g.mu must be held for writing.
func (g *liveGraph) rebuild() {
	newState := &GraphState{}
	for _, object := range g.objects {
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

	if g.metadataStale {
		if fresh, err := g.queryMetadata(); err == nil {
			newState.Metadatas = fresh
			g.metadataStale = false
		}
	}
	g.state = newState
}
