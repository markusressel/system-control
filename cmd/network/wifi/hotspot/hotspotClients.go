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
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "List currently connected Hotspot Client Devices",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		isHotspotUp, err := wifi.IsHotspotUp(hotspotSSID)
		if err != nil {
			return err
		}
		if !isHotspotUp {
			fmt.Println("Hotspot is not running")
			return nil
		}

		hotspotDevices, err := wifi.GetConnectedHotspotDevices(wifiInterface, hotspotSSID)
		if err != nil {
			return err
		}

		for _, device := range hotspotDevices {
			printHotspotDevice(device)
		}

		return err
	},
}

func init() {
	Command.AddCommand(clientsCmd)

	clientsCmd.PersistentFlags().StringVarP(
		&hotspotSSID,
		"ssid", "s",
		"",
		"SSID of the hotspot",
	)
}

func printHotspotDevice(device wifi.HotspotLease) {
	properties := orderedmap.NewOrderedMap[string, string]()
	properties.Set("IP", device.IP)
	properties.Set("MAC", device.MAC)

	util.PrintFormattedTableOrdered(device.Name, properties)
}
