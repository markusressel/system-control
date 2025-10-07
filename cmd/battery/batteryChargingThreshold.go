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
	"strconv"

	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var batteryChargingThresholdCmd = &cobra.Command{
	Use:   "threshold",
	Short: "Get/Set battery charging threshold",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		if len(args) > 0 {
			newValueString := args[0]
			newValue, err := strconv.Atoi(newValueString)
			if err != nil {
				return err
			}
			err = setBatteryThreshold(battery, newValue)
			if err != nil {
				return err
			}
		} else {
			value, err := getBatteryThreshold(battery)
			if err != nil {
				return err
			}
			fmt.Println(value)
		}

		return nil
	},
}

var batteryChargingThresholdSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current battery charging threshold on disk",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		current, err := getBatteryThreshold(battery)
		if err != nil {
			return err
		}

		return persistence.SaveInt(battery+"_charge_control_end_threshold", current)
	},
}

var batteryChargingThresholdRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the last saved battery charging threshold",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		value, err := persistence.ReadInt(battery + "_charge_control_end_threshold")
		if err != nil {
			return err
		}
		return setBatteryThreshold(battery, int(value))
	},
}

func setBatteryThreshold(battery string, value int) error {
	file := "/sys/class/power_supply/" + battery + "/charge_control_end_threshold"
	return util.WriteIntToFile(value, file)
}

func getBatteryThreshold(battery string) (int, error) {
	file := "/sys/class/power_supply/" + battery + "/charge_control_end_threshold"

	value, err := util.ReadIntFromFile(file)
	if err != nil {
		return -1, err
	}
	return int(value), err
}

func init() {
	Command.AddCommand(batteryChargingThresholdCmd)

	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdSaveCmd)
	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdRestoreCmd)
}
