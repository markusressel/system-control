package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var positionCmd = &cobra.Command{
	Use:   "position",
	Short: "Show playback position",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.position"),
}

func init() {
	positionCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
	Command.AddCommand(positionCmd)
}
