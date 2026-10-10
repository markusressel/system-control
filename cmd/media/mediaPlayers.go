package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var playersCmd = &cobra.Command{
	Use:   "players",
	Short: "List available players",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.players"),
}

func init() {
	Command.AddCommand(playersCmd)
}
