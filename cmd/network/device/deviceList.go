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
package device

import (
	"cmp"
	"fmt"
	orderedmap "github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
	"slices"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List network devices",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		networkDevices, err := wifi.GetNetworkDevices()
		if err != nil {
			return err
		}

		// sort entries
		slices.SortFunc(networkDevices, func(a, b wifi.NetworkDevice) int {
			return cmp.Or(
				// sort by connected state
				cmp.Compare(a.State, b.State),
				// then sort by type
				cmp.Compare(a.Type, b.Type),
				// then sort by name
				util.CompareIgnoreCase(a.Name, b.Name),
			)
		})

		// type NetworkDevice struct {
		//	Name            string
		//	Type            string
		//	State           string
		//	IP4Connectivity string // IPv4 connectivity
		//	IP6Connectivity string // IPv6 connectivity
		//	DBUSPath        string // DBUS path
		//	Connection      string // Connection name
		//	CONUUID            string // CONUUID
		//	CONPath         string // Connection path
		//}

		for i, device := range networkDevices {
			properties := orderedmap.NewOrderedMap[string, string]()
			properties.Set("Name", device.Name)
			properties.Set("Type", device.Type)
			properties.Set("State", device.State)
			properties.Set("IPv4-Connectivity", device.IP4Connectivity)
			properties.Set("IPv6-Connectivity", device.IP6Connectivity)
			properties.Set("DBus-Path", device.DBUSPath)
			properties.Set("Connection", device.Connection)
			properties.Set("Con-UUID", device.CONUUID)
			properties.Set("Con-Path", device.CONPath)

			util.PrintFormattedTableOrdered(device.Name, properties)

			if i < len(networkDevices)-1 {
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(listCmd)
}
