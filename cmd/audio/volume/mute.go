package volume

import (
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var muteCmd = &cobra.Command{
	Use:   "mute",
	Short: "Mute system audio",
	Long:  ``,
	RunE: commandutil.LockedRunE("cmd-audio-volume", commandutil.MutatorRunE(func(cmd *cobra.Command, args []string) ([]commandutil.MutatorResult, error) {
		state := pipewire.PwDump()

		targets, err := resolveTargets(state, device, stream)
		if err != nil {
			return nil, err
		}

		var results []commandutil.MutatorResult
		for _, target := range targets {
			err = pipewire.WpCtlSetMute(target.Id, true)
			if err != nil {
				return nil, err
			}
			results = append(results, commandutil.MutatorResult{
				CommandId: "audio.muted",
				Flags:     nil,
				Value:     "yes",
			})
		}

		return results, nil
	})),
}

func init() {
	VolumeCmd.AddCommand(muteCmd)
}
