package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var setVolumeCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a specific volume",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  commandutil.OpRunE("audio.volume.set"),
}

func init() {
	VolumeCmd.AddCommand(setVolumeCmd)
}
