/*
 * system-control
 * Copyright (c) 2019. Markus Ressel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
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
