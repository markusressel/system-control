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
		if len(args) == 0 || args[0] == "" {
			return bluetooth.DisconnectAllBluetoothDevices()
		}

		devices, err := bluetooth.GetBluetoothDevices()
		if err != nil {
			return err
		}

		device, err := resolveDevice(args[0], devices)
		if err != nil {
			return err
		}
		return bluetooth.DisconnectBluetoothDevice(device)
	},
}

func init() {
	Command.AddCommand(bluetoothDisconnectCmd)
}
