package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var device string
var stream string

var VolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Show the current volume",
	RunE:  commandutil.OpRunE("audio.volume"),
}

func init() {
	VolumeCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")

	VolumeCmd.PersistentFlags().StringVarP(
		&device,
		"device", "d",
		"",
		"Device Name/Description",
	)

	VolumeCmd.PersistentFlags().StringVarP(
		&stream,
		"stream", "s",
		"",
		"Stream Name/Description",
	)
}
