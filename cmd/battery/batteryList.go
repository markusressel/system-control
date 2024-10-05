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
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
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
		for i, battery := range batteries {
			properties := map[string]string{
				"Path":           battery.Path,
				"Type":           battery.Type,
				"Manufacturer":   battery.Manufacturer,
				"Model":          battery.Model,
				"Serial":         battery.SerialNumber,
				"Capacity":       strconv.Itoa(int(battery.Capacity)),
				"Capacity Level": battery.CapacityLevel,
				"Online":         strconv.FormatBool(battery.Online),
				"Status":         battery.Status,
				"Scope":          battery.Scope,
			}

			util.PrintFormattedTable(battery.Name, properties)

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
