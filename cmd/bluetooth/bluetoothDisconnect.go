package bluetooth

import (
	"fmt"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect a currently connected Bluetooth Device",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceName := ""
		if len(args) > 0 {
			deviceName = args[0]
		}

		if deviceName == "" {
			return bluetooth.DisconnectAllBluetoothDevices()
		} else {
			devices, err := bluetooth.GetBluetoothDevices()
			if err != nil {
				return err
			}
			for _, device := range devices {
				if device.Name == deviceName || device.Address == deviceName {
					err := bluetooth.DisconnectBluetoothDevice(device)
					return err
				}
			}
		}

		return fmt.Errorf("device not found: %v", deviceName)
	},
}

func init() {
	Command.AddCommand(bluetoothDisconnectCmd)
}
