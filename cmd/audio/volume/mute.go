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
	"strconv"
)

var muteCmd = &cobra.Command{
	Use:   "mute",
	Short: "Mute system audio",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nameFlag := cmd.Flag("name")
		name := nameFlag.Value.String()

		var targetSink map[string]string
		if name == "" {
			targetSink = pipewire.GetActiveSinkPipewire()
		} else {
			targetSink = pipewire.GetSinkByName(name)
		}
		targetSinkDeviceId, err := strconv.Atoi(targetSink["device.id"])
		if err != nil {
			return err
		}
		return pipewire.SetMutedPipewire(targetSinkDeviceId, true)
	},
}

func init() {
	VolumeCmd.AddCommand(muteCmd)
}
