package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var togglePlayPauseCmd = &cobra.Command{
	Use:     "togglePlayPause",
	Aliases: []string{"toggle-play-pause"},
	Short:   "Toggle play/pause",
	Args:    cobra.NoArgs,
	RunE:    commandutil.OpRunE("media.togglePlayPause"),
}

func init() {
	Command.AddCommand(togglePlayPauseCmd)
}
