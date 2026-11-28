package pipewire

import "testing"

// An update must not change the info it is merged into: that info is shared with snapshots
// that other goroutines read without a lock.
func TestParseInfoMerge_LeavesTheExistingInfoAlone(t *testing.T) {
	existing := InterfaceNodeInfo{
		State:  "running",
		Props:  map[string]interface{}{"node.name": "alsa_output", "media.class": "Audio/Sink"},
		Params: map[string]interface{}{"Props": []interface{}{map[string]interface{}{"volume": 0.5}}},
	}

	merged := parseInfoMerge(TypeNode, []byte(`{"props": {"node.name": "renamed"}, "params": {"Props": [{"volume": 0.7}]}}`),
		true, existing).(InterfaceNodeInfo)

	if existing.Props["node.name"] != "alsa_output" {
		t.Errorf("existing props changed: %v", existing.Props)
	}
	if volume := existing.Params["Props"].([]interface{})[0].(map[string]interface{})["volume"]; volume != 0.5 {
		t.Errorf("existing params changed: volume %v", volume)
	}

	if merged.Props["node.name"] != "renamed" || merged.Props["media.class"] != "Audio/Sink" || merged.State != "running" {
		t.Errorf("update not merged: %+v", merged)
	}
	if volume := merged.Params["Props"].([]interface{})[0].(map[string]interface{})["volume"]; volume != 0.7 {
		t.Errorf("merged volume = %v, want 0.7", volume)
	}
}

func TestParseInfoMerge_NewProfiler(t *testing.T) {
	info := parseInfoMerge(TypeProfiler, []byte(`{"a": 1}`), false, nil).(InterfaceProfilerInfo)
	if info == nil || (*info)["a"] != 1.0 {
		t.Errorf("profiler info = %v", info)
	}
}

// An incremental update only carries what changed, the rest of the object stays.
func TestMergeObject_KeepsWhatTheUpdateLeavesOut(t *testing.T) {
	existing, _ := mergeObject(RawGraphObject{
		Id: 42, Type: TypeNode, Version: 3, Permissions: []string{"r"},
		Props: map[string]interface{}{"object.serial": 7.0},
		Info:  []byte(`{"state": "idle", "props": {"node.name": "sink"}}`),
	}, GraphObject{}, false)

	merged, stale := mergeObject(RawGraphObject{Id: 42, Info: []byte(`{"state": "running"}`)}, existing, true)

	if stale {
		t.Error("an update without metadata is stale")
	}
	if merged.Type != TypeNode || merged.Version != 3 || len(merged.Permissions) != 1 || merged.Props["object.serial"] != 7.0 {
		t.Errorf("common data not kept: %+v", merged.CommonData)
	}
	info := merged.Info.(InterfaceNodeInfo)
	if info.State != "running" || info.Props["node.name"] != "sink" {
		t.Errorf("info = %+v", info)
	}
}

func TestIsRemoval(t *testing.T) {
	if !isRemoval(RawGraphObject{Id: 1, Info: []byte("null")}) || !isRemoval(RawGraphObject{Id: 1}) {
		t.Error("a removal was not recognized")
	}
	if isRemoval(RawGraphObject{Id: 1, Type: TypeNode}) {
		t.Error("a new object without info is no removal")
	}
}
