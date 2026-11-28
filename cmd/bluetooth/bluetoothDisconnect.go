package bluetooth

import (
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect a currently connected Bluetooth Device",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		return bluetooth.DisconnectByName(name)
	},
}

func init() {
	Command.AddCommand(bluetoothDisconnectCmd)
}
