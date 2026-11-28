package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var VolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Show the current volume",
	RunE:  commandutil.OpRunE("audio.volume"),
}

func init() {
	commandutil.AddWatchFlag(VolumeCmd)

	VolumeCmd.PersistentFlags().StringP(
		"device", "d",
		"",
		"Device Name/Description",
	)

	VolumeCmd.PersistentFlags().StringP(
		"stream", "s",
		"",
		"Stream Name/Description",
	)
}
