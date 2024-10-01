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
		for _, device := range devices {
			if device.Name == deviceName || device.Address == deviceName {
				err := bluetooth.ConnectToBluetoothDevice(device)
				return err
			}
		}

		return fmt.Errorf("device not found: %v", deviceName)
	},
}

func init() {
	Command.AddCommand(bluetoothConnectCmd)
}
