package volume

import (
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var toggleMuteCmd = &cobra.Command{
	Use:   "toggle-mute",
	Short: "Toggle the Mute state",
	Long:  ``,
	RunE: commandutil.LockedRunE("cmd-audio-volume", commandutil.MutatorRunE(func(cmd *cobra.Command, args []string) ([]commandutil.MutatorResult, error) {
		state := pipewire.PwDump()

		targets, err := resolveTargets(state, device, stream)
		if err != nil {
			return nil, err
		}

		var results []commandutil.MutatorResult
		for _, target := range targets {
			initialMuted, _ := state.IsMuted(target.Id)

			err = pipewire.WpCtlToggleMute(target.Id)
			if err != nil {
				return nil, err
			}

			val := "yes"
			if initialMuted {
				val = "no"
			}
			results = append(results, commandutil.MutatorResult{
				CommandId: "audio.muted",
				Flags:     nil,
				Value:     val,
			})
		}

		return results, nil
	})),
}

func init() {
	VolumeCmd.AddCommand(toggleMuteCmd)
}
