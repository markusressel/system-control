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

			bPresent, e := battery.IsPresent()
			bPresentText := ""
			if e == nil {
				bPresentText = strconv.FormatBool(bPresent)
			}
			properties.Set("Present", bPresentText)

			properties.Set("Path", battery.Path)
			bType, _ := battery.GetType()
			properties.Set("Type", bType)
			properties.Set("Manufacturer", battery.Manufacturer)
			properties.Set("Model", battery.Model)
			properties.Set("Serial", battery.SerialNumber)

			bCapacity, e := battery.GetCapacity()
			bCapacityText := ""
			if e == nil {
				bCapacityText = strconv.Itoa(int(bCapacity))
				bCapacityText = fmt.Sprintf("%v %%", bCapacityText)
			}
			properties.Set("Capacity", bCapacityText)

			bCapacityLevel, _ := battery.GetCapacityLevel()
			properties.Set("Capacity Level", bCapacityLevel)

			bCycleCount, e := battery.GetCycleCount()
			bCycleCountText := ""
			if e == nil {
				bCycleCountText = strconv.Itoa(int(bCycleCount))
			}
			properties.Set("Cycle Count", bCycleCountText)

			bEnergyFullDesign, e := battery.GetEnergyFullDesign()
			bEnergyFullDesignText := ""
			if e == nil {
				bEnergyFullDesignText = fmt.Sprintf("%v Wh", util.RoundToTwoDecimals(bEnergyFullDesign))
			}
			properties.Set("Energy Full Design", bEnergyFullDesignText)

			bEnergyFull, e := battery.GetEnergyFull()
			bEnergyFullText := ""
			if e == nil {
				bEnergyFullText = fmt.Sprintf("%v Wh", util.RoundToTwoDecimals(bEnergyFull))
			}
			properties.Set("Energy Full", bEnergyFullText)

			bEnergyNow, e := battery.GetEnergyNow()
			bEnergyNowText := ""
			if e == nil {
				bEnergyNowText = fmt.Sprintf("%v Wh", util.RoundToTwoDecimals(bEnergyNow))
			}
			properties.Set("Energy Now", bEnergyNowText)

			degradation, e := battery.GetDegradation()
			degradationText := ""
			if e == nil {
				degradationText = fmt.Sprintf("%v %%", util.RoundToTwoDecimals(degradation))
			}
			properties.Set("Degradation", degradationText)

			bPowerNow, e := battery.GetPowerNow()
			bPowerNowText := ""
			if e == nil {
				bPowerNowText = fmt.Sprintf("%v W", util.RoundToTwoDecimals(bPowerNow))
			}
			properties.Set("Power Now", bPowerNowText)

			bOnline, e := battery.IsOnline()
			bOnlineText := ""
			if e == nil {
				bOnlineText = strconv.FormatBool(bOnline)
			}
			properties.Set("Online", bOnlineText)

			bStatus, _ := battery.GetStatus()
			properties.Set("Status", bStatus)
			properties.Set("Scope", battery.Scope)
			bTechnology, _ := battery.GetTechnology()
			properties.Set("Technology", bTechnology)

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
