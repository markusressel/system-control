package pipewire

import "testing"

// The default source has its own metadata keys, next to the ones of the default sink.
func TestGetDefaultSourceNode(t *testing.T) {
	speaker := sinkNode(1, "alsa_output.speaker")
	mic := sinkNode(2, "alsa_input.mic")
	headsetMic := sinkNode(3, "bluez_input.AA")
	state := GraphState{
		Nodes: []InterfaceNode{speaker, mic, headsetMic},
		Metadatas: []InterfaceMetadata{{
			CommonData: CommonData{Props: map[string]interface{}{"metadata.name": "default"}, Metadata: []map[string]interface{}{
				{"key": "default.configured.audio.sink", "value": name("alsa_output.speaker")},
				// the configured headset is switched off, WirePlumber uses the microphone
				{"key": "default.configured.audio.source", "value": name("bluez_input.BB")},
				{"key": "default.audio.source", "value": name("alsa_input.mic")},
			}},
		}},
	}

	source, err := state.GetDefaultSourceNode()
	if err != nil {
		t.Fatal(err)
	}
	if source.Id != 2 {
		t.Errorf("source = %d, want 2", source.Id)
	}

	sink, err := state.GetDefaultSinkNode()
	if err != nil {
		t.Fatal(err)
	}
	if sink.Id != 1 {
		t.Errorf("sink = %d, want 1", sink.Id)
	}
}

func TestGetInputStreamNodes(t *testing.T) {
	recording := sinkNode(1, "recording")
	recording.Info.Props["media.class"] = "Stream/Input/Audio"
	playback := sinkNode(2, "playback")
	playback.Info.Props["media.class"] = "Stream/Output/Audio"
	state := GraphState{Nodes: []InterfaceNode{recording, playback}}

	streams := state.GetInputStreamNodes()

	if len(streams) != 1 || streams[0].Id != 1 {
		t.Errorf("input streams = %v, want only the recording", streams)
	}
}
