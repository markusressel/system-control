package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	internalmedia "github.com/markusressel/system-control/internal/media"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:              "media",
	Short:            "Control media players via playerctl",
	TraverseChildren: true,
	RunE:             commandutil.OpRunE("media"),
}

func init() {
	commandutil.AddWatchFlag(Command)
	Command.PersistentFlags().StringP(
		internalmedia.PlayerFlag, "p",
		"",
		"Player name or regex",
	)
}
