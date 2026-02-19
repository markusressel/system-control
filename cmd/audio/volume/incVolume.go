package volume

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var IncVolumeCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increment audio volume",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		state := pipewire.PwDump()

		volume, err := state.GetVolumeByName(device)
		volume = util.RoundToTwoDecimals(volume)
		if err != nil {
			return err
		}
		change := audio.CalculateAppropriateVolumeChange(volume*100, true) / 100.0

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

		targetVolume := volume + change
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
	VolumeCmd.AddCommand(IncVolumeCmd)
}
