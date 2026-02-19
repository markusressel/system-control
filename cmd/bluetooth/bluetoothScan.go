package bluetooth

import (
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var bluetoothScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for available Bluetooth Devices",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			err = bluetooth.SetBluetoothScan(false)
			if err != nil {
				println("Failed to turn off Bluetooth scanning")
			}
		}()
		err = bluetooth.SetBluetoothScan(true)
		if err != nil {
			return err
		}

		println("Scanning for Bluetooth Devices...")

		devices, err := bluetooth.GetBluetoothDevices()
		if err != nil {
			return err
		}
		availableDevices := util.FilterFunc(devices, func(device bluetooth.BluetoothDevice) bool {
			return !device.Connected && !device.Paired
		})

		printBluetoothDevices(availableDevices)

		return nil
	},
}

func init() {
	Command.AddCommand(bluetoothScanCmd)
}
