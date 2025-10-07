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
package device

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get/Set the current profile of a device",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		state := pipewire.PwDump()

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		}

		device, err := state.FindDeviceByName(deviceName)
		if err != nil {
			return err
		}

		if len(profileName) > 0 {
			profile, err := device.GetProfileIdByName(profileName)
			if err != nil {
				return err
			}

			return device.SetProfileByName(profile.Name)
		} else {
			profile, err := device.GetActiveProfile()
			if err != nil {
				return err
			}
			fmt.Printf(profile.Description)
			return nil
		}
	},
}

func init() {
	DeviceCmd.AddCommand(profileCmd)
}
