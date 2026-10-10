package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause media",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.pause"),
}

func init() {
	Command.AddCommand(pauseCmd)
}
