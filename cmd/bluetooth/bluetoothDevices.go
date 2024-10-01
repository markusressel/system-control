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
	"strconv"
)

var filterConnected bool
var filterPaired bool

var bluetoothDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List all known Bluetooth Devices",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		devices, err := bluetooth.GetBluetoothDevices()
		if err != nil {
			return err
		}

		for _, device := range devices {
			if filterConnected && !device.Connected {
				continue
			}
			if filterPaired && !device.Paired {
				continue
			}

			fmt.Printf("Name: %v\n", device.Name)
			fmt.Printf("Address: %v\n", device.Address)
			fmt.Printf("Connected: %v\n", device.Connected)
			fmt.Printf("Paired: %v\n", device.Paired)
			if device.BatteryPercentage != nil {
				fmt.Printf("Battery: %v%%\n", *device.BatteryPercentage)
			}
			fmt.Println()
		}

		return nil
	},
}

func parseFlagAsBool(cmd *cobra.Command, flagName string, defaultValue bool) (flagValue bool, err error) {
	connectedFlag := cmd.Flag(flagName)
	flagValue = defaultValue
	if connectedFlag != nil {
		connectedFlagValue := connectedFlag.Value.String()
		if connectedFlagValue != "" {
			flagValue, err = strconv.ParseBool(connectedFlagValue)
			if err != nil {
				return flagValue, err
			}
		}
	}
	return flagValue, nil
}

func init() {
	Command.AddCommand(bluetoothDevicesCmd)

	bluetoothDevicesCmd.Flags().BoolVarP(&filterConnected, "connected", "c", false, "Filter by connected state")
	bluetoothDevicesCmd.Flags().BoolVarP(&filterPaired, "paired", "p", false, "Filter by paired state")
}
