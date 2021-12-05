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
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strconv"
)

// batteryChargingThresholdCmd represents the batteryChargingThreshold command
var batteryChargingThresholdCmd = &cobra.Command{
	Use:   "threshold",
	Short: "Get/Set battery charging threshold",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		path := "/sys/class/power_supply/" + battery + "/charge_control_end_threshold"

		if len(args) > 0 {
			newValueString := args[0]
			newValue, err := strconv.Atoi(newValueString)
			if err != nil {
				log.Fatal(err)
			}
			err = writeIntToFile(newValue, path)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			value, err := readIntFromFile(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(value)
		}
	},
}

func init() {
	batteryCmd.AddCommand(batteryChargingThresholdCmd)
}
