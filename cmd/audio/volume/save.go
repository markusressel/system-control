package volume

import (
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/spf13/cobra"
)

type audioState struct {
	Device string
	Volume float64
	Muted  bool
}

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current state of the given audio device",
	Long:  ``,
	RunE: commandutil.LockedRunE("cmd-audio-volume", func(cmd *cobra.Command, args []string) error {
		state := pipewire.PwDump()
		targets, err := resolveTargets(state, device, stream)
		if err != nil {
			return err
		}

		for _, target := range targets {
			name, _ := target.GetName()
			volume, _ := state.GetVolumeByName(name)
			muted, _ := state.IsMuted(target.Id)

			key := computeKey(name)
			data := audioState{
				Device: name,
				Volume: volume,
				Muted:  muted,
			}
			err := persistence.SaveStruct(key, &data)
			if err != nil {
				return err
			}
		}

		return nil
	}),
}

func computeKey(deviceName string) string {
	return "audio_state_" + deviceName
}

func init() {
	VolumeCmd.AddCommand(saveCmd)
}
