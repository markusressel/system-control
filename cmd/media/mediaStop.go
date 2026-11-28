package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop media",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.stop"),
}

func init() {
	Command.AddCommand(stopCmd)
}
