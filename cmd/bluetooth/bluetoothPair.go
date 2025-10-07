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
	"time"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var autoConnect bool
var removeExisting bool

var bluetoothPairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Pair a Bluetooth Device",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceName := args[0]

		if removeExisting {
			devices, err := bluetooth.GetBluetoothDevices()
			if err != nil {
				return err
			}
			for _, device := range devices {
				if device.Name == deviceName || device.Address == deviceName {
					err := bluetooth.RemoveBluetoothDevice(device)
					if err != nil {
						return fmt.Errorf("failed to remove existing device %s: %v", deviceName, err)
					}
				}
			}
		}

		defer func() {
			err := bluetooth.SetBluetoothScan(false)
			if err != nil {
				fmt.Println(err.Error())
			}
		}()

		err := bluetooth.SetBluetoothScan(true)
		if err != nil {
			return fmt.Errorf("failed to start Bluetooth scan: %v", err)
		}

		startTime := time.Now()
		// search for device in loop with timeout
		for time.Now().Sub(startTime) < 30*time.Second {
			devices, err := bluetooth.GetBluetoothDevices()
			if err != nil {
				return err
			}
			for _, device := range devices {
				if device.Name == deviceName || device.Address == deviceName {
					err := bluetooth.PairBluetoothDevice(device)

					if err != nil && autoConnect {
						fmt.Printf("Failed to pair device %s: %v\n", deviceName, err)
					} else if autoConnect {
						err = bluetooth.ConnectToBluetoothDevice(device)
						if err != nil {
							fmt.Printf("Failed to connect to device %s after pairing: %v\n", deviceName, err)
						}
					}
					return err
				}
			}

			time.Sleep(1 * time.Second)
		}

		return fmt.Errorf("device not found: %v", deviceName)
	},
}

func init() {
	Command.AddCommand(bluetoothPairCmd)
	bluetoothPairCmd.Flags().BoolVarP(&autoConnect, "connect", "c", false, "Automatically connect after pairing")
	bluetoothPairCmd.Flags().BoolVarP(&removeExisting, "remove-existing", "r", false, "Remove existing device (if it exists) before pairing")
}
