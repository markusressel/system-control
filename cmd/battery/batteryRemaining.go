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
	"math"
)

var batteryRemainingCmd = &cobra.Command{
	Use:   "remaining",
	Short: "Get the remaining battery life",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		template := "%hours%:%minutes%"

		battery := Name

		// get value
		charging, err := util.IsBatteryCharging(battery)
		if err != nil {
			return err
		}
		energyTarget, err := util.GetEnergyTarget(battery)
		if err != nil {
			return err
		}
		energyNow, err := util.GetEnergyNow(battery)
		if err != nil {
			return err
		}
		powerNow, err := util.GetPowerNow(battery)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		if powerNow == 0 {
			fmt.Printf("∞")
			return nil
		}

		var remainingTimeInSeconds int64
		if charging == true {
			remainingTimeInSeconds = util.CalculateRemainingTime(energyTarget-energyNow, powerNow)
		} else {
			remainingTimeInSeconds = util.CalculateRemainingTime(energyNow, powerNow)
		}

		remainingHours := int(math.Min(99, float64(remainingTimeInSeconds/60/60)))
		remainingMinutes := (remainingTimeInSeconds / 60) % 60

		placeholders := map[string]string{}
		placeholders["hours"] = fmt.Sprintf("%02d", remainingHours)
		placeholders["minutes"] = fmt.Sprintf("%02d", remainingMinutes)

		result := util.ReplacePlaceholders(template, placeholders)
		fmt.Printf(result)

		return nil
	},
}

func init() {
	Command.AddCommand(batteryRemainingCmd)
}
