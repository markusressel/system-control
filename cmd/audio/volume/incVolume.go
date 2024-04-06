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
package volume

import (
	"fmt"
	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var IncVolumeCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increment audio volume",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		nameFlag := cmd.Flag("name")
		name := nameFlag.Value.String()

		state := pipewire.PwDump()

		volume, err := state.GetVolumeByName(name)
		volume = util.RoundToTwoDecimals(volume)
		if err != nil {
			return err
		}
		change := audio.CalculateAppropriateVolumeChange(volume*100, true) / 100.0

		var target pipewire.InterfaceNode
		if name == "" {
			target, err = state.GetDefaultNode()
		} else {
			target, err = state.GetNodeByName(name)
		}
		if err != nil {
			return err
		}

		parentDevice, err := target.GetParentDevice()
		if err != nil {
			return err
		}

		targetVolume := volume + change
		err = state.SetVolume(parentDevice.Id, targetVolume)
		if err != nil {
			return err
		}

		newVolume, err := state.GetVolumeByName(name)
		if err != nil {
			return err
		}
		newVolume = util.RoundToTwoDecimals(newVolume)
		volumeAsInt := (int)(newVolume * 100)
		fmt.Println(volumeAsInt)
		return err
	},
}

func init() {
	VolumeCmd.AddCommand(IncVolumeCmd)
}
