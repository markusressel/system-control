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
package backlight

import (
	"strconv"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var setBrightnessCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the brightness of a given display backlight.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		mainBacklight, err := util.GetMainBacklight()
		if err != nil {
			return err
		}
		return mainBacklight.SetBrightness(p)
	},
}

func init() {
	brightnessCmd.AddCommand(setBrightnessCmd)
}
