package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var previousCmd = &cobra.Command{
	Use:   "previous",
	Short: "Go to previous track",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.previous"),
}

func init() {
	Command.AddCommand(previousCmd)
}
