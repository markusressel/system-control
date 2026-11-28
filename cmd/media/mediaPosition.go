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
	commandutil.AddWatchFlag(positionCmd)
	Command.AddCommand(positionCmd)
}
