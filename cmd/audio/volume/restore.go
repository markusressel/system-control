package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the state of the given audio device from a previous save",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.volume.restore"),
}

func init() {
	VolumeCmd.AddCommand(restoreCmd)
}
