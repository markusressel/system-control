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
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"sort"
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
			return fmt.Errorf("multiple matching devices found: %v", matchingDevices)
		}

		return fmt.Errorf("device not found: %v", deviceName)
	},
}

func findBluetoothDeviceFuzzy(name string, devices []bluetooth.BluetoothDevice) []bluetooth.BluetoothDevice {
	// check exact address matches first
	for _, device := range devices {
		if util.EqualsIgnoreCase(device.Address, name) {
			return []bluetooth.BluetoothDevice{device}
		}
	}

	// then check fuzzy name matches
	deviceNames := make([]string, len(devices))
	for i, device := range devices {
		deviceNames[i] = device.Name
	}

	fuzzyMatches := fuzzy.RankFindNormalizedFold(name, deviceNames)
	sort.Sort(fuzzyMatches)

	result := make([]bluetooth.BluetoothDevice, 0)
	for _, match := range fuzzyMatches {
		for _, device := range devices {
			if device.Name == match.Target {
				result = append(result, device)
			}
		}
	}

	return result
}

func init() {
	Command.AddCommand(bluetoothConnectCmd)
}
