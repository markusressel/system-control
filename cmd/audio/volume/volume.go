package volume

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var device string
var stream string

func resolveTargets(state pipewire.GraphState, targetDevice string, targetStream string) ([]pipewire.InterfaceNode, error) {
	var targets []pipewire.InterfaceNode
	if targetStream != "" {
		targets = state.FindStreamNodes(targetStream)
	} else if targetDevice != "" {
		targets = state.FindNodesByName(targetDevice)
	} else {
		target, err := state.GetDefaultSinkNode()
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

var VolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Show the current volume",
	RunE: commandutil.StateCommandRunE("audio.volume", nil, func(cmd *cobra.Command) (string, error) {
		state := pipewire.PwDump()

		volume, err := state.GetVolumeByName(device)
		if err != nil {
			return "", err
		}
		volume = util.RoundToTwoDecimals(volume)
		volumeAsInt := (int)(volume * 100)
		return fmt.Sprintf("%d", volumeAsInt), nil
	}),
}

func init() {
	VolumeCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")

	VolumeCmd.PersistentFlags().StringVarP(
		&device,
		"device", "d",
		"",
		"Device Name/Description",
	)

	VolumeCmd.PersistentFlags().StringVarP(
		&stream,
		"stream", "s",
		"",
		"Stream Name/Description",
	)
}
