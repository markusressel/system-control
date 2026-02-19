package touchpad

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var touchpadOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn off the Touchpad",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		util.SetTouchpadEnabled(false)
	},
}

func init() {
	Command.AddCommand(touchpadOffCmd)
}
