package battery

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var batteryRemainingCmd = &cobra.Command{
	Use:   "remaining",
	Short: "Get the remaining battery life in hours and minutes",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("battery.remaining"),
}

func init() {
	batteryRemainingCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
	Command.AddCommand(batteryRemainingCmd)
}
