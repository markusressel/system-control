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
	"github.com/markusressel/system-control/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
	path2 "path"
	"strconv"
)

var batteryChargingThresholdCmd = &cobra.Command{
	Use:   "threshold",
	Short: "Get/Set battery charging threshold",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		if len(args) > 0 {
			newValueString := args[0]
			newValue, err := strconv.Atoi(newValueString)
			if err != nil {
				log.Fatal(err)
			}
			setBatteryThreshold(battery, newValue)
		} else {
			value := getBatteryThreshold(battery)
			fmt.Println(value)
		}
	},
}

var batteryChargingThresholdSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current battery charging threshold on disk",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		current := getBatteryThreshold(battery)

		err := os.MkdirAll(internal.ConfigBaseDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		path := path2.Join(internal.ConfigBaseDir, battery+"_charge_control_end_threshold.sav")
		err = internal.WriteIntToFile(current, path)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var batteryChargingThresholdRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the last saved battery charging threshold",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		path := path2.Join(internal.ConfigBaseDir, battery+"_charge_control_end_threshold.sav")
		value, err := internal.ReadIntFromFile(path)
		if err != nil {
			return
		}
		setBatteryThreshold(battery, int(value))
	},
}

func setBatteryThreshold(battery string, value int) {
	path := "/sys/class/power_supply/" + battery + "/charge_control_end_threshold"
	err := internal.WriteIntToFile(value, path)
	if err != nil {
		log.Fatal(err)
	}
}

func getBatteryThreshold(battery string) int {
	path := "/sys/class/power_supply/" + battery + "/charge_control_end_threshold"

	value, err := internal.ReadIntFromFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return int(value)
}

func init() {
	Command.AddCommand(batteryChargingThresholdCmd)

	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdSaveCmd)
	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdRestoreCmd)
}
