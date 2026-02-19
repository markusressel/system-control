package volume

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

//var device string
//var stream string

var mutedCmd = &cobra.Command{
	Use:   "muted",
	Short: "Show the current mute state",
	RunE: func(cmd *cobra.Command, args []string) error {
		state := pipewire.PwDump()

		node, err := state.GetDefaultSinkNode()
		if err != nil {
			return err
		}
		muted, err := state.IsMuted(node.Id)
		if err != nil {
			return err
		}
		if muted {
			fmt.Println("yes")
		} else {
			fmt.Println("no")
		}
		return nil
	},
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

	VolumeCmd.AddCommand(mutedCmd)
}
