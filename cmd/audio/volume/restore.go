package volume

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the state of the given audio device from a previous save",
	Long:  ``,
	RunE: commandutil.LockedRunE("cmd-audio-volume", commandutil.MutatorRunE(func(cmd *cobra.Command, args []string) ([]commandutil.MutatorResult, error) {
		state := pipewire.PwDump()
		targets, err := resolveTargets(state, device, stream)
		if err != nil {
			return nil, err
		}

		var results []commandutil.MutatorResult
		for _, target := range targets {
			name, _ := target.GetName()
			key := computeKey(name)
			data := audioState{}
			err := persistence.ReadStruct(key, &data)
			if err != nil {
				continue // No saved state
			}

			err = pipewire.WpCtlSetMute(target.Id, data.Muted)
			if err != nil {
				return nil, err
			}
			err = pipewire.WpCtlSetVolume(target.Id, data.Volume)
			if err != nil {
				return nil, err
			}

			val := "yes"
			if !data.Muted {
				val = "no"
			}
			newVolume := util.RoundToTwoDecimals(data.Volume)
			volumeAsInt := (int)(newVolume * 100)

			results = append(results, commandutil.MutatorResult{
				CommandId: "audio.muted",
				Flags:     map[string]string{"device": device},
				Value:     val,
			})
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
	VolumeCmd.AddCommand(restoreCmd)
}
