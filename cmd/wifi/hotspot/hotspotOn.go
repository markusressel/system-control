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
package wifi

import (
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var onCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn on the WiFi Hotspot",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: get this from somewhere
		hotspotName := "M16 Hotspot"
		err := wifi.TurnOnHotspot(
			hotspotName,
			"wlo1",
			"M16",
			"M12345678",
		)
		return err
	},
}

func init() {
	Command.AddCommand(onCmd)
}
