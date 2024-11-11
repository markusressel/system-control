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
package battery

import (
	"cmp"
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"slices"
	"strconv"
)

var batteryListCmd = &cobra.Command{
	Use:   "list",
	Short: "Get a list of all known batteries",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		batteries, err := util.GetBatteryList()
		if err != nil {
			return err
		}

		// sort entries
		slices.SortFunc(batteries, func(a, b util.BatteryInfo) int {
			return cmp.Or(
				// sort by battery name
				util.CompareIgnoreCase(a.Name, b.Name),
			)
		})

		for i, battery := range batteries {
			properties := orderedmap.NewOrderedMap[string, string]()
			properties.Set("Path", battery.Path)
			properties.Set("Type", battery.Type)
			properties.Set("Manufacturer", battery.Manufacturer)
			properties.Set("Model", battery.Model)
			properties.Set("Serial", battery.SerialNumber)
			properties.Set("Capacity", strconv.Itoa(int(battery.Capacity)))
			properties.Set("Capacity Level", battery.CapacityLevel)
			properties.Set("Online", strconv.FormatBool(battery.Online))
			properties.Set("Status", battery.Status)
			properties.Set("Scope", battery.Scope)

			util.PrintFormattedTableOrdered(battery.Name, properties)

			if i < len(batteries)-1 {
				fmt.Println()
			}

		}
		return nil
	},
}

func init() {
	Command.AddCommand(batteryListCmd)
}
