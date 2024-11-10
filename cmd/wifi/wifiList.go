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
	orderedmap "github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
	"strconv"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known WiFi networks",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		networks, err := wifi.GetNetworks()
		if err != nil {
			return err
		}

		for i, network := range networks {
			properties := orderedmap.NewOrderedMap[string, string]()
			properties.Set("Connected", strconv.FormatBool(network.Connected))
			properties.Set("SSID", network.SSID)
			properties.Set("BSSID", network.BSSID)
			properties.Set("Mode", network.Mode)
			properties.Set("Channel", fmt.Sprintf("%v", network.Channel))
			properties.Set("Bandwidth", fmt.Sprintf("%v", network.Bandwidth))
			properties.Set("Frequency", fmt.Sprintf("%v", network.Frequency))
			properties.Set("Rate", fmt.Sprintf("%v", network.Rate))
			properties.Set("Signal", fmt.Sprintf("%v", network.Signal))
			properties.Set("Bars", fmt.Sprintf("%v", network.Bars))
			properties.Set("Security", network.Security)

			util.PrintFormattedTableOrdered(network.SSID, properties)

			if i < len(networks)-1 {
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(listCmd)
}
