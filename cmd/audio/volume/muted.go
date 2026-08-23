package volume

import (
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

//var device string
//var stream string

var mutedCmd = &cobra.Command{
	Use:   "muted",
	Short: "Show the current mute state",
	RunE: commandutil.StateCommandRunE("audio.muted", nil, func(cmd *cobra.Command) (string, error) {
		state := pipewire.PwDump()

		node, err := state.GetDefaultSinkNode()
		if err != nil {
			return "", err
		}
		muted, err := state.IsMuted(node.Id)
		if err != nil {
			return "", err
		}
		if muted {
			return "yes", nil
		} else {
			return "no", nil
		}
	}),
}

func init() {
	//mutedCmd.PersistentFlags().StringVarP(
	//	&device,
	//	"device", "d",
	//	"",
	//	"Device Name/Description",
	//)
	//
	//mutedCmd.PersistentFlags().StringVarP(
	//	&stream,
	//	"stream", "s",
	//	"",
	//	"Stream Name/Description",
	//)

	mutedCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")

	VolumeCmd.AddCommand(mutedCmd)
}
