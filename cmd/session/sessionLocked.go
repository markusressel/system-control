package session

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var lockedCmd = &cobra.Command{
	Use:   "locked",
	Short: "Check whether the desktop session is locked",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("session.locked"),
}

func init() {
	lockedCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
	Command.AddCommand(lockedCmd)
}
