package bluetooth

import (
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var scanDuration time.Duration

var bluetoothScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for available Bluetooth Devices",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		err = bluetooth.SetBluetoothScan(true)
		if err != nil {
			return err
		}
		defer func() {
			if scanErr := bluetooth.SetBluetoothScan(false); scanErr != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "Failed to turn off Bluetooth scanning")
			}
		}()

		fmt.Fprintln(cmd.ErrOrStderr(), "Scanning for Bluetooth Devices...")
		time.Sleep(scanDuration)

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

	bluetoothScanCmd.Flags().DurationVar(&scanDuration, "duration", 5*time.Second, "How long to scan for devices")
}
