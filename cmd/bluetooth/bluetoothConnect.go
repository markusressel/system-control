package bluetooth

import (
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a Bluetooth Device",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		devices, err := bluetooth.GetBluetoothDevices()
		if err != nil {
			return err
		}

		device, err := resolveDevice(args[0], devices)
		if err != nil {
			return err
		}
		return bluetooth.ConnectToBluetoothDevice(device)
	},
}

func init() {
	Command.AddCommand(bluetoothConnectCmd)
}
