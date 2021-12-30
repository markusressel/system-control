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
package display

import (
	"github.com/markusressel/system-control/internal"
	"github.com/spf13/cobra"
)

var incBrightnessCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase display brightness",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		brightness := internal.GetBrightness()
		maxBrightness := internal.GetMaxBrightness()

		percentage := int((float32(brightness) / float32(maxBrightness)) * 100.0)

		var change int
		if percentage < 10 {
			change = 1
		} else if percentage < 20 {
			change = 2
		} else if percentage < 40 {
			change = 4
		} else {
			change = 8
		}

		rawChange := int(float32(change) * (float32(maxBrightness) / 100.0))

		internal.AdjustBrightness(rawChange)
	},
}

func init() {
	brightnessCmd.AddCommand(incBrightnessCmd)
}
