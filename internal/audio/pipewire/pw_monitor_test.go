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
