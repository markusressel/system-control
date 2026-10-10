package touchpad

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var setTouchpadCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the enabled state of the Touchpad",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  commandutil.OpRunE("touchpad.set"),
}

func init() {
	Command.AddCommand(setTouchpadCmd)
}
