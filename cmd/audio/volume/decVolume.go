package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var decVolumeCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrement audio volume",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.volume.dec"),
}

func init() {
	VolumeCmd.AddCommand(decVolumeCmd)
}
