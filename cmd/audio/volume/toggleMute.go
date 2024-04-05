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
	"strconv"
)

var toggleMuteCmd = &cobra.Command{
	Use:   "toggle-mute",
	Short: "Toggle the Mute state",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		nameFlag := cmd.Flag("name")
		name := nameFlag.Value.String()

		var targetSink map[string]string
		if name == "" {
			targetSink = pipewire.GetActiveSinkPipewire()
		} else {
			targetSink = pipewire.GetSinkByName(name)
		}
		sinkId, err := strconv.Atoi(targetSink["id"])
		targetSinkDeviceId, err := strconv.Atoi(targetSink["device.id"])
		isMuted := pipewire.IsMutedPipewire(sinkId)
		if err != nil {
			return err
		}
		err = pipewire.SetMutedPipewire(targetSinkDeviceId, !isMuted)
		fmt.Println(sinkId)
		return nil
	},
}

func init() {
	VolumeCmd.AddCommand(toggleMuteCmd)
}
