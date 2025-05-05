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
package mouse

import (
	"github.com/markusressel/system-control/internal/upower"
	"github.com/spf13/cobra"
)

var getMouseBatteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Get the current battery level",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		battery, err := GetMouseBattery()
		if err != nil {
			return err
		}
		cmd.Println(battery)
		return nil
	},
}

func GetMouseBattery() (string, error) {
	upowerDevices, err := upower.GetUpowerDevices()
	if err != nil {
		return "", err
	}
	for _, device := range upowerDevices {
		if device.Type == "mouse" {
			return device.Percentage, nil
		}
	}
	return "", nil
}

func init() {
	Command.AddCommand(getMouseBatteryCmd)
}
