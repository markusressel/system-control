package touchpad

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var toggleTouchpadCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle the Touchpad state",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		isTouchpadEnabled, err := util.IsTouchpadEnabled()
		if err != nil {
			return err
		}
		return util.SetTouchpadEnabled(!isTouchpadEnabled)
	},
}

func init() {
	Command.AddCommand(toggleTouchpadCmd)
}
