package touchpad

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var toggleTouchpadVerticalScrollDirection = &cobra.Command{
	Use:   "toggleVerticalScrollDirection",
	Short: "Toggle the vertical scroll direction of the Touchpad",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.ToggleTouchpadScrollVerticalDirection()
	},
}

func init() {
	Command.AddCommand(toggleTouchpadVerticalScrollDirection)
}
