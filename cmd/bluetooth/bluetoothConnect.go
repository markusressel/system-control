package bluetooth

import (
	"fmt"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a Bluetooth Device",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceName := args[0]

		devices, err := bluetooth.GetBluetoothDevices()
		if err != nil {
			return err
		}

		matchingDevices := findBluetoothDeviceFuzzy(deviceName, devices)

		if len(matchingDevices) == 1 {
			err := bluetooth.ConnectToBluetoothDevice(matchingDevices[0])
			if err != nil {
				return err
			}
			return nil
		} else if len(matchingDevices) > 1 {
			deviceNames := createDeviceNameList(matchingDevices)
			return fmt.Errorf("multiple matching devices found: %v", deviceNames)
		}

		return fmt.Errorf("device not found: %v", deviceName)
	},
}

func init() {
	Command.AddCommand(bluetoothConnectCmd)
}
