package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to next track",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.next"),
}

func init() {
	Command.AddCommand(nextCmd)
}
