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
package cmd

import (
	"github.com/spf13/cobra"
)

// incVolumeCmd represents the inc command
var incVolumeCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increment audio volume",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		channelFlag := cmd.Flag("channel")
		channel := channelFlag.Value.String()
		volume := getVolume(channel)
		setVolume(channel, volume+1)
	},
}

func init() {
	volumeCmd.AddCommand(incVolumeCmd)

	// Here you will define your flags and configuration settings.

}
