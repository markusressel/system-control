package mouse

import (
	"github.com/markusressel/system-control/internal/commandutil"
	internalmouse "github.com/markusressel/system-control/internal/mouse"
	"github.com/spf13/cobra"
)

var getMouseBatteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Get the current battery level",
	Long:  ``,
	RunE:  commandutil.OpRunE("mouse.battery"),
}

func init() {
	commandutil.AddWatchFlag(getMouseBatteryCmd)
	getMouseBatteryCmd.Flags().String(internalmouse.BatteryTypeFlag, internalmouse.BatteryTypePercentage, "output format: percentage or number")
	Command.AddCommand(getMouseBatteryCmd)
}
