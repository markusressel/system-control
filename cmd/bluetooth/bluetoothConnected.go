package bluetooth

import (
	"fmt"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothConnectedCmd = &cobra.Command{
	Use:   "connected",
	Short: "Check if currently connected to the given Bluetooth Device",
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
			if matchingDevices[0].Connected == true {
				fmt.Println("yes")
				return nil
			}
		} else if len(matchingDevices) > 1 {
			deviceNames := createDeviceNameList(matchingDevices)
			return fmt.Errorf("multiple matching devices found: %v", deviceNames)
		}

		fmt.Println("no")
		return nil
	},
}

func init() {
	Command.AddCommand(bluetoothConnectedCmd)
}
