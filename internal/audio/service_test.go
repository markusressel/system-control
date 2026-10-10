package audio

import (
	"math"
	"testing"
	"time"

	"github.com/markusressel/system-control/internal/audio/pipewire"
)

func fakeNode(id int, name string, volume float64, muted bool) pipewire.InterfaceNode {
	return pipewire.InterfaceNode{
		CommonData: pipewire.CommonData{Id: id},
		Info: pipewire.InterfaceNodeInfo{
			Props: map[string]interface{}{"node.name": name},
			Params: map[string]interface{}{
				"Props": []interface{}{
					map[string]interface{}{
						"channelVolumes": []interface{}{math.Pow(volume, 3)},
						"mute":           muted,
					},
				},
			},
		},
	}
}

func fakeGraph(defaultSink string, nodes ...pipewire.InterfaceNode) pipewire.GraphState {
	return pipewire.GraphState{
		Nodes: nodes,
		Metadatas: []pipewire.InterfaceMetadata{{
			CommonData: pipewire.CommonData{
				Props: map[string]interface{}{"metadata.name": "default"},
				Metadata: []map[string]interface{}{{
					"key":   "default.audio.sink",
					"value": map[string]interface{}{"name": defaultSink},
				}},
			},
		}},
	}
}

type setCall struct {
	id     int
	volume float64
}

func newFakeService(graph *pipewire.GraphState, calls *[]setCall) *Service {
	return NewService(Backend{
		Graph: func() pipewire.GraphState { return *graph },
		SetVolume: func(id int, volume float64) error {
			*calls = append(*calls, setCall{id, volume})
			return nil
		},
		SetMute:    func(id int, mute bool) error { return nil },
		ToggleMute: func(id int) error { return nil },
	})
}

func TestResolveTargets_SpecificDevice(t *testing.T) {
	state := fakeGraph("TestDevice", fakeNode(1, "TestDevice", 0.5, false))

	targets, err := ResolveTargets(state, "TestDevice", "")
	if err != nil {
		t.Fatalf("ResolveTargets returned error: %v", err)
	}
	if len(targets) != 1 || targets[0].Id != 1 {
		t.Fatalf("expected exactly target 1, got %v", targets)
	}
}

func TestSavedStateKey(t *testing.T) {
	if key := savedStateKey("my_device"); key != "audio_state_my_device" {
		t.Errorf("savedStateKey = %s", key)
	}
}

// Rapid steps must build on the volume written by the previous step, even though
// the graph still reports the old value.
func TestStep_BuildsOnPreviousWriteWhileGraphIsStale(t *testing.T) {
	graph := fakeGraph("sinkA", fakeNode(7, "sinkA", 0.50, false))
	var calls []setCall
	s := newFakeService(&graph, &calls)

	for i := 0; i < 2; i++ {
		if _, err := s.Step("", "", true); err != nil {
			t.Fatal(err)
		}
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(calls))
	}
	if math.Abs(calls[0].volume-0.55) > 1e-9 || math.Abs(calls[1].volume-0.60) > 1e-9 {
		t.Errorf("expected 0.55 then 0.60, got %v then %v", calls[0].volume, calls[1].volume)
	}
}

// The remembered volume is dropped when it is old or belongs to another sink.
func TestStep_IgnoresStaleOptimisticState(t *testing.T) {
	graph := fakeGraph("sinkA", fakeNode(7, "sinkA", 0.50, false), fakeNode(8, "sinkB", 0.20, false))
	var calls []setCall
	s := newFakeService(&graph, &calls)

	if _, err := s.Step("", "", true); err != nil {
		t.Fatal(err)
	}

	// default sink switches: the remembered 0.55 must not be applied to sink B
	graph = fakeGraph("sinkB", graph.Nodes...)
	if _, err := s.Step("", "", true); err != nil {
		t.Fatal(err)
	}
	if got := calls[1]; got.id != 8 || math.Abs(got.volume-0.22) > 1e-9 {
		t.Errorf("expected sink 8 at 0.22, got %+v", got)
	}

	// an external change after the window must be picked up
	graph = fakeGraph("sinkB", fakeNode(7, "sinkA", 0.50, false), fakeNode(8, "sinkB", 0.80, false))
	s.lastVolume[8] = written[float64]{0.22, time.Now().Add(-2 * optimisticWindow)}
	if _, err := s.Step("", "", false); err != nil {
		t.Fatal(err)
	}
	if got := calls[2]; math.Abs(got.volume-0.75) > 1e-9 {
		t.Errorf("expected 0.75 based on the external change, got %v", got.volume)
	}
}

func TestStep_ClampsAtZero(t *testing.T) {
	graph := fakeGraph("sinkA", fakeNode(7, "sinkA", 0.00, false))
	var calls []setCall
	s := newFakeService(&graph, &calls)

	lines, err := s.Step("", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if calls[0].volume != 0 || lines[0] != "0" {
		t.Errorf("expected clamp to 0, got %v / %v", calls[0].volume, lines)
	}
}

func TestMute_ReportsResultingState(t *testing.T) {
	graph := fakeGraph("sinkA", fakeNode(7, "sinkA", 0.5, true))
	var calls []setCall
	s := newFakeService(&graph, &calls)

	lines, err := s.Mute("", "", MuteToggle)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "no" {
		t.Errorf("toggling a muted sink must report \"no\", got %v", lines)
	}
}

// A query right after a change must report the new value although the graph still has the old one.
func TestVolumeAndMuted_ReflectOwnRecentWrites(t *testing.T) {
	graph := fakeGraph("sinkA", fakeNode(7, "sinkA", 0.50, false))
	var calls []setCall
	s := newFakeService(&graph, &calls)

	if v, _ := s.Volume(""); v != "50" {
		t.Fatalf("volume before = %s", v)
	}
	if _, err := s.Step("", "", true); err != nil {
		t.Fatal(err)
	}
	if v, _ := s.Volume(""); v != "55" {
		t.Errorf("volume right after a step = %s, want 55", v)
	}

	if _, err := s.Mute("", "", MuteOn); err != nil {
		t.Fatal(err)
	}
	if m, _ := s.Muted("", ""); m != "yes" {
		t.Errorf("muted right after muting = %s, want yes", m)
	}

	// once the window passed the graph is the source of truth again
	s.lastVolume[7] = written[float64]{0.55, time.Now().Add(-2 * optimisticWindow)}
	s.lastMuted[7] = written[bool]{true, time.Now().Add(-2 * optimisticWindow)}
	if v, _ := s.Volume(""); v != "50" {
		t.Errorf("volume after the window = %s, want 50", v)
	}
	if m, _ := s.Muted("", ""); m != "no" {
		t.Errorf("muted after the window = %s, want no", m)
	}
}

// Two toggles in quick succession must end up in the original state although the graph lags.
func TestMute_ToggleTwiceWhileGraphIsStale(t *testing.T) {
	graph := fakeGraph("sinkA", fakeNode(7, "sinkA", 0.5, false))
	var calls []setCall
	s := newFakeService(&graph, &calls)

	first, err := s.Mute("", "", MuteToggle)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Mute("", "", MuteToggle)
	if err != nil {
		t.Fatal(err)
	}
	if first[0] != "yes" || second[0] != "no" {
		t.Errorf("toggle results = %v, %v; want yes, no", first, second)
	}
}
