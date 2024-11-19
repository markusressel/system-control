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
package wifi

import (
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
	"os"
)

var force bool

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn off the WiFi Hotspot",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wifiInterface == "" {
			wifiInterface = "wlo1"
		}
		if hotspotSSID == "" {
			hotspotSSID = "M16"
		}

		hotspotName := createHotspotConfigName(hotspotSSID)

		if !force {
			// check if there are still devices connected
			hotspotDevices, err := wifi.GetConnectedHotspotDevices(wifiInterface, hotspotSSID)
			if err != nil {
				return err
			}

			if len(hotspotDevices) > 0 {
				for _, device := range hotspotDevices {
					printHotspotDevice(device)
				}
				println("There are still devices connected to the hotspot. Please disconnect them first or use --force parameter.")
				os.Exit(1)
			}
		}

		err := wifi.TurnOffHotspot(hotspotName)
		return err
	},
}

func init() {
	Command.AddCommand(offCmd)

	offCmd.Flags().BoolVarP(
		&force,
		"force", "f",
		false,
		"Force turn off the Hotspot",
	)

	offCmd.PersistentFlags().StringVarP(
		&hotspotSSID,
		"ssid", "s",
		"",
		"SSID of the hotspot",
	)
}
