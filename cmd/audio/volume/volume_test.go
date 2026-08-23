package volume

import (
	"testing"

	"github.com/markusressel/system-control/internal/audio/pipewire"
)

func TestComputeKey(t *testing.T) {
	expected := "audio_state_my_device"
	if key := computeKey("my_device"); key != expected {
		t.Errorf("computeKey(\"my_device\") = %s; want %s", key, expected)
	}
}

func TestVolumeCommandsRegistered(t *testing.T) {
	expectedCmds := []string{"inc", "dec", "mute", "unmute", "toggle-mute", "save", "restore"}

	for _, expected := range expectedCmds {
		found := false
		for _, cmd := range VolumeCmd.Commands() {
			if cmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' to be registered under VolumeCmd, but it was not", expected)
		}
	}
}

func TestResolveTargets_SpecificDevice(t *testing.T) {
	// Create a mock state
	state := pipewire.GraphState{
		Nodes: []pipewire.InterfaceNode{
			{
				CommonData: pipewire.CommonData{Id: 1},
				Info: pipewire.InterfaceNodeInfo{
					Props: map[string]interface{}{
						"node.name": "TestDevice",
					},
				},
			},
		},
	}

	targets, err := resolveTargets(state, "TestDevice", "")
	if err != nil {
		t.Fatalf("resolveTargets returned error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}

	if targets[0].Id != 1 {
		t.Errorf("expected target Id 1, got %d", targets[0].Id)
	}
}

func TestResolveTargets_SpecificStream(t *testing.T) {
	// Create a mock state
	state := pipewire.GraphState{
		Nodes: []pipewire.InterfaceNode{
			{
				CommonData: pipewire.CommonData{Id: 2},
				Info: pipewire.InterfaceNodeInfo{
					Props: map[string]interface{}{
						"node.name":   "TestStream",
						"media.class": "Stream/Output/Audio",
					},
				},
			},
		},
	}

	targets, err := resolveTargets(state, "", "TestStream")
	if err != nil {
		t.Fatalf("resolveTargets returned error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}

	if targets[0].Id != 2 {
		t.Errorf("expected target Id 2, got %d", targets[0].Id)
	}
}
