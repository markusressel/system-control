package sink

import (
	"errors"
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the default sink",
	Long: `Switches the default audio sink and moves all existing audio streams to the given one.
You can specify the audio sink using its index, but also using other strings that occur in its description:

> system-control audio sink switch "headphone"

> system-control audio sink switch "NVIDIA"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		searchString := args[0]
		//sinkIdx := findSinkPulse(searchString)
		//switchSinkPulse(sinkIdx)

		state := pipewire.PwDump()

		nodes := state.FindNodesByName(searchString)
		audioSinkNodes := filterByMediaClass(nodes, pipewire.MediaClassAudioSink)

		if len(audioSinkNodes) <= 0 {
			return errors.New("no sink found")
		}
		if len(audioSinkNodes) > 1 {
			nodeNames := make([]string, len(nodes))
			for i, node := range nodes {
				name, _ := node.GetName()
				description, _ := node.GetDescription()
				nodeNames[i] = fmt.Sprintf("%s (%s)", name, description)
			}

			return errors.New(fmt.Sprintf("ambiguous sink name, found: %v", nodeNames))
		}
		node := audioSinkNodes[0]
		return state.SwitchSinkTo(node)
	},
}

func filterByMediaClass(nodes []pipewire.InterfaceNode, mediaClass string) []pipewire.InterfaceNode {
	result := make([]pipewire.InterfaceNode, 0)
	for _, node := range nodes {
		nodeMediaClass, _ := node.GetMediaClass()
		if nodeMediaClass == mediaClass {
			result = append(result, node)
		}
	}
	return result
}

func init() {
	SinkCmd.AddCommand(switchCmd)
}
