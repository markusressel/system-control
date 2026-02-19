package volume

import (
	"fmt"
	"strconv"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var setVolumeCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a specific volume",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volume, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		targetVolume := float64(volume)
		targetVolume = float64(volume) / 100.0

		state := pipewire.PwDump()

		var targets []pipewire.InterfaceNode
		if stream != "" {
			targets = state.FindStreamNodes(stream)
		} else if device != "" {
			targets = state.FindNodesByName(device)
		} else {
			target, err := state.GetDefaultSinkNode()
			if err != nil {
				return err
			}
			targets = append(targets, target)
		}

		for _, target := range targets {
			err = pipewire.WpCtlSetVolume(target.Id, targetVolume)
			if err != nil {
				return err
			}

			state = pipewire.PwDump()
			newVolume, err := state.GetVolumeByName(device)
			if err != nil {
				return err
			}
			newVolume = util.RoundToTwoDecimals(newVolume)
			volumeAsInt := (int)(newVolume * 100)
			fmt.Println(volumeAsInt)
		}

		return err
	},
}

func init() {
	VolumeCmd.AddCommand(setVolumeCmd)
}
