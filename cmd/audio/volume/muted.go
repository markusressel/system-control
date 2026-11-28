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
	commandutil.AddWatchFlag(mutedCmd)

	VolumeCmd.AddCommand(mutedCmd)
}
