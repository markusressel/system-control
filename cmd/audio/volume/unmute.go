package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var unmuteCmd = &cobra.Command{
	Use:   "unmute",
	Short: "Unmute system audio",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.volume.unmute"),
}

func init() {
	VolumeCmd.AddCommand(unmuteCmd)
}
