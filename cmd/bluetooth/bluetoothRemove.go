package bluetooth

import (
	"fmt"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Bluetooth Device",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceName := ""
		if len(args) > 0 {
			deviceName = args[0]
		}

		devices, err := bluetooth.GetBluetoothDevices()
		if err != nil {
			return err
		}
		for _, device := range devices {
			if device.Name == deviceName || device.Address == deviceName {
				err := bluetooth.RemoveBluetoothDevice(device)
				return err
			}
		}

		return fmt.Errorf("device not found: %v", deviceName)
	},
}

func init() {
	Command.AddCommand(bluetoothRemoveCmd)
}
