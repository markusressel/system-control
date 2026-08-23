package volume

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var IncVolumeCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increment audio volume",
	Long:  ``,
	RunE: commandutil.LockedRunE("cmd-audio-volume", commandutil.MutatorRunE(func(cmd *cobra.Command, args []string) ([]commandutil.MutatorResult, error) {
		state := pipewire.PwDump()

		volume, err := state.GetVolumeByName(device)
		volume = util.RoundToTwoDecimals(volume)
		if err != nil {
			return nil, err
		}
		change := audio.CalculateAppropriateVolumeChange(volume*100, true) / 100.0

		targets, err := resolveTargets(state, device, stream)
		if err != nil {
			return nil, err
		}

		targetVolume := volume + change
		var results []commandutil.MutatorResult
		for _, target := range targets {
			err = pipewire.WpCtlSetVolume(target.Id, targetVolume)
			if err != nil {
				return nil, err
			}

			newVolume := util.RoundToTwoDecimals(targetVolume)
			volumeAsInt := (int)(newVolume * 100)

			results = append(results, commandutil.MutatorResult{
				CommandId: "audio.volume",
				Flags:     map[string]string{"device": device},
				Value:     fmt.Sprintf("%d", volumeAsInt),
			})
		}
		return results, nil
	})),
}

func init() {
	VolumeCmd.AddCommand(IncVolumeCmd)
}
