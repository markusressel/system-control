package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show media status",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.status"),
}

func init() {
	commandutil.AddWatchFlag(statusCmd)
	Command.AddCommand(statusCmd)
}
