package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var IncVolumeCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increment audio volume",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.volume.inc"),
}

func init() {
	VolumeCmd.AddCommand(IncVolumeCmd)
}
