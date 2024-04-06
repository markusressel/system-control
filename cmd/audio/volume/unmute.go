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

var unmuteCmd = &cobra.Command{
	Use:   "unmute",
	Short: "Unmute system audio",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		state := pipewire.PwDump()

		var target pipewire.InterfaceNode
		if stream != "" {
			target, err = state.GetStreamNode(stream)
		} else if device != "" {
			target, err = state.GetNodeByName(device)
		} else {
			target, err = state.GetDefaultSinkNode()
		}
		if err != nil {
			return err
		}

		return pipewire.WpCtlSetMute(target.Id, false)
	},
}

func init() {
	VolumeCmd.AddCommand(unmuteCmd)
}
