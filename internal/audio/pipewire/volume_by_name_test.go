package pipewire

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// a sound card with a line out (profile device 5) and a microphone (profile device 0), like the onboard audio of a PC
func soundCardGraph() GraphState {
	route := func(direction string, profileDevice int, linearVolume float64) DeviceRoute {
		return DeviceRoute{
			Direction: direction,
			Device:    profileDevice,
			Props:     map[string]interface{}{"channelVolumes": []interface{}{linearVolume, linearVolume}},
		}
	}
	node := func(id int, name string, profileDevice int) InterfaceNode {
		return InterfaceNode{
			CommonData: CommonData{Id: id},
			Info: InterfaceNodeInfo{Props: map[string]interface{}{
				"node.name":           name,
				"device.id":           float64(91),
				"card.profile.device": float64(profileDevice),
			}},
		}
	}
	return GraphState{
		Devices: []InterfaceDevice{{
			CommonData: CommonData{Id: 91},
			Info: InterfaceDeviceInfo{Params: DeviceInfoParams{Route: []DeviceRoute{
				// linear volumes, 0.064 is 0.4 and 0.216 is 0.6 on the cubic scale
				route("Input", 0, 0.064),
				route("Output", 5, 0.216),
			}}},
		}},
		Nodes: []InterfaceNode{
			node(35, "alsa_output.pci-0000_11_00.4.analog-stereo", 5),
			node(98, "alsa_input.pci-0000_11_00.4.analog-stereo", 0),
		},
	}
}

func TestGetVolumeByNameUsesTheRouteOfTheNode(t *testing.T) {
	state := soundCardGraph()

	for _, test := range []struct {
		name     string
		expected float64
	}{
		{"alsa_output.pci-0000_11_00.4.analog-stereo", 0.6},
		{"alsa_input.pci-0000_11_00.4.analog-stereo", 0.4},
	} {
		t.Run(test.name, func(t *testing.T) {
			volume, err := state.GetVolumeByName(test.name)

			assert.NoError(t, err)
			assert.InDelta(t, test.expected, volume, 0.0001)
		})
	}
}
