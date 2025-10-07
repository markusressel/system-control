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
package keyboard

import (
	"fmt"
	"strconv"

	"github.com/markusressel/system-control/internal/util"

	"github.com/spf13/cobra"
)

var keyboardBrightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Get/Set the current keyboard backlight brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			p, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			_, err = util.SetKeyboardBrightness(p)
			if err != nil {
				return err
			}
		} else {
			brightness, err := util.GetKeyboardBrightness()
			if err != nil {
				return err
			}
			fmt.Println(brightness)
		}

		return nil
	},
}

func init() {
	Command.AddCommand(keyboardBrightnessCmd)
}
