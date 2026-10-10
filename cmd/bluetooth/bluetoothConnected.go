package bluetooth

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var bluetoothConnectedCmd = &cobra.Command{
	Use:   "connected",
	Short: "Check if currently connected to the given Bluetooth Device",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  commandutil.OpRunE("bluetooth.connected"),
}

func init() {
	Command.AddCommand(bluetoothConnectedCmd)

	bluetoothConnectedCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
}
