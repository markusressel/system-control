package volume

import (
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var toggleMuteCmd = &cobra.Command{
	Use:   "toggle-mute",
	Short: "Toggle the Mute state",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
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
			err = pipewire.WpCtlToggleMute(target.Id)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	VolumeCmd.AddCommand(toggleMuteCmd)
}
