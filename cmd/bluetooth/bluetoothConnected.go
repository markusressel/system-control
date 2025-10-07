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
