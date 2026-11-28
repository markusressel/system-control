package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var toggleMuteCmd = &cobra.Command{
	Use:   "toggle-mute",
	Short: "Toggle the Mute state",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.volume.toggle-mute"),
}

func init() {
	VolumeCmd.AddCommand(toggleMuteCmd)
}
