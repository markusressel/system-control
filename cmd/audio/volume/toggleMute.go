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
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var toggleMuteCmd = &cobra.Command{
	Use:   "toggle-mute",
	Short: "Toggle the Mute state",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		state := pipewire.PwDump()

		var target pipewire.InterfaceNode
		if device == "" {
			target, err = state.GetDefaultNode()
		} else {
			target, err = state.GetNodeByName(device)
		}
		if err != nil {
			return err
		}

		parentDevice, err := target.GetParentDevice()
		if err != nil {
			return err
		}

		sinkId := target.Id
		isMuted, err := state.IsMuted(sinkId)
		if err != nil {
			return err
		}
		err = state.SetMuted(parentDevice.Id, !isMuted)
		return nil
	},
}

func init() {
	VolumeCmd.AddCommand(toggleMuteCmd)
}
