package touchpad

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var toggleTouchpadVerticalScrollDirection = &cobra.Command{
	Use:   "toggleVerticalScrollDirection",
	Short: "Toggle the vertical scroll direction of the Touchpad",
	Long:  ``,
	RunE:  commandutil.OpRunE("touchpad.toggleVerticalScrollDirection"),
}

func init() {
	Command.AddCommand(toggleTouchpadVerticalScrollDirection)
}
