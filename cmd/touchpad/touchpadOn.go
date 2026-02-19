package touchpad

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var touchpadOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn on the Touchpad",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		util.SetTouchpadEnabled(true)
	},
}

func init() {
	Command.AddCommand(touchpadOnCmd)
}
