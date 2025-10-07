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

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

//var device string
//var stream string

var mutedCmd = &cobra.Command{
	Use:   "muted",
	Short: "Show the current mute state",
	RunE: func(cmd *cobra.Command, args []string) error {
		state := pipewire.PwDump()

		node, err := state.GetDefaultSinkNode()
		if err != nil {
			return err
		}
		muted, err := state.IsMuted(node.Id)
		if err != nil {
			return err
		}
		if muted {
			fmt.Println("yes")
		} else {
			fmt.Println("no")
		}
		return nil
	},
}

func init() {
	//mutedCmd.PersistentFlags().StringVarP(
	//	&device,
	//	"device", "d",
	//	"",
	//	"Device Name/Description",
	//)
	//
	//mutedCmd.PersistentFlags().StringVarP(
	//	&stream,
	//	"stream", "s",
	//	"",
	//	"Stream Name/Description",
	//)

	VolumeCmd.AddCommand(mutedCmd)
}
