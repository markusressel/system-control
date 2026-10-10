package touchpad

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var toggleTouchpadCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle the Touchpad state",
	Long:  ``,
	RunE:  commandutil.OpRunE("touchpad.toggle"),
}

func init() {
	Command.AddCommand(toggleTouchpadCmd)
}
