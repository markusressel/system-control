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
	"fmt"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var backlightListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current display brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		backlights, err := util.GetBacklights()
		if err != nil {
			return err
		}

		for i, backlight := range backlights {
			brightness, _ := backlight.GetBrightness()
			maxBrightness, _ := backlight.GetMaxBrightness()

			properties := orderedmap.NewOrderedMap[string, string]()
			properties.Set("Brightness", fmt.Sprintf("%d", brightness))
			properties.Set("MaxBrightness", fmt.Sprintf("%d", maxBrightness))

			util.PrintFormattedTableOrdered(backlight.Name, properties)

			if i < len(backlights)-1 {
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(backlightListCmd)
}
