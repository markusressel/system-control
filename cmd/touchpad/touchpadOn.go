package touchpad

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var touchpadOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn on the Touchpad",
	Long:  ``,
	RunE:  commandutil.OpRunE("touchpad.on"),
}

func init() {
	Command.AddCommand(touchpadOnCmd)
}
