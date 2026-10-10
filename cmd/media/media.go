package media

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:              "media",
	Short:            "Control media players via playerctl",
	TraverseChildren: true,
	RunE:             commandutil.OpRunE("media"),
}

func init() {
	Command.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
	Command.PersistentFlags().StringP(
		"player", "p",
		"",
		"Player name or regex",
	)
}
