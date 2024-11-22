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
	"github.com/elliotchance/orderedmap/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"sort"
	"strconv"
)

var Name string

var Command = &cobra.Command{
	Use:              "bluetooth",
	Short:            "Control Bluetooth Devices",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	Command.PersistentFlags().StringVarP(
		&Name,
		"name", "n",
		"",
		"Device Name",
	)
}

func printBluetoothDevices(devices []bluetooth.BluetoothDevice) {
	for i, device := range devices {
		printBluetoothDevice(device)
		if i < len(devices)-1 {
			fmt.Println()
		}
	}
}

func printBluetoothDevice(device bluetooth.BluetoothDevice) {
	properties := orderedmap.NewOrderedMap[string, string]()
	properties.Set("Address", device.Address)
	properties.Set("Connected", strconv.FormatBool(device.Connected))
	properties.Set("Paired", strconv.FormatBool(device.Paired))

	if device.BatteryPercentage != nil {
		properties.Set("Battery", fmt.Sprintf("%v%%", *device.BatteryPercentage))
	}

	util.PrintFormattedTableOrdered(device.Name, properties)
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

func createDeviceNameList(devices []bluetooth.BluetoothDevice) []string {
	return util.MapFunc(devices, func(device bluetooth.BluetoothDevice) string {
		return fmt.Sprintf("%s (%s)", device.Name, device.Address)
	})
}
