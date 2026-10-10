package volume

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var mutedCmd = &cobra.Command{
	Use:   "muted",
	Short: "Show the current mute state",
	RunE:  commandutil.OpRunE("audio.volume.muted"),
}

func init() {
	mutedCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")

	VolumeCmd.AddCommand(mutedCmd)
}
