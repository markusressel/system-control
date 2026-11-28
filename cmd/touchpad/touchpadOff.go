package touchpad

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var touchpadOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn off the Touchpad",
	Long:  ``,
	RunE:  commandutil.OpRunE("touchpad.off"),
}

func init() {
	Command.AddCommand(touchpadOffCmd)
}
