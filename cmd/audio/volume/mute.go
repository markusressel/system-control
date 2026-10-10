package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var muteCmd = &cobra.Command{
	Use:   "mute",
	Short: "Mute system audio",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.volume.mute"),
}

func init() {
	VolumeCmd.AddCommand(muteCmd)
}
