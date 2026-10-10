package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play media",
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("media.play"),
}

func init() {
	Command.AddCommand(playCmd)
}
