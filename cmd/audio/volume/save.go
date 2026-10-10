package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current state of the given audio device",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.volume.save"),
}

func init() {
	VolumeCmd.AddCommand(saveCmd)
}
