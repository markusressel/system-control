package pipewire

import "testing"

func sinkNode(id int, name string) InterfaceNode {
	return InterfaceNode{
		CommonData: CommonData{Id: id},
		Info:       InterfaceNodeInfo{Props: map[string]interface{}{"node.name": name}},
	}
}

func graphWithDefaults(configured, active interface{}, nodes ...InterfaceNode) GraphState {
	entries := []map[string]interface{}{}
	if configured != nil {
		entries = append(entries, map[string]interface{}{"key": "default.configured.audio.sink", "value": configured})
	}
	if active != nil {
		entries = append(entries, map[string]interface{}{"key": "default.audio.sink", "value": active})
	}
	return GraphState{
		Nodes: nodes,
		Metadatas: []InterfaceMetadata{{
			CommonData: CommonData{Props: map[string]interface{}{"metadata.name": "default"}, Metadata: entries},
		}},
	}
}

func name(n string) map[string]interface{} { return map[string]interface{}{"name": n} }

func TestGetDefaultSinkNode(t *testing.T) {
	headset := sinkNode(1, "bluez_output.AA.1")
	speaker := sinkNode(2, "alsa_output.speaker")

	tests := []struct {
		name   string
		state  GraphState
		wantId int
		fail   bool
	}{
		{"configured and present", graphWithDefaults(name(headset.Info.Props["node.name"].(string)), name("alsa_output.speaker"), headset, speaker), 1, false},
		{"configured sink is switched off, the active one is used", graphWithDefaults(name("bluez_output.AA.1"), name("alsa_output.speaker"), speaker), 2, false},
		{"empty configured sink", graphWithDefaults(name(""), name("alsa_output.speaker"), speaker), 2, false},
		{"json string values", graphWithDefaults(`{"name":"alsa_output.speaker"}`, nil, speaker), 2, false},
		{"exact name wins over a longer name", graphWithDefaults(name("sink.1"), nil, sinkNode(3, "sink.10"), sinkNode(4, "sink.1")), 4, false},
		{"neither is present", graphWithDefaults(name("gone"), name("also.gone"), speaker), 0, true},
		{"no defaults at all", graphWithDefaults(nil, nil, speaker), 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node, err := tc.state.GetDefaultSinkNode()
			if tc.fail {
				if err == nil {
					t.Fatalf("expected an error, got node %d", node.Id)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if node.Id != tc.wantId {
				t.Errorf("node = %d, want %d", node.Id, tc.wantId)
			}
		})
	}
}
