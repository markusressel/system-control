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
	"fmt"
	"github.com/spf13/cobra"
)

// volumeCmd represents the volume command
var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Show the current volume",
	Run: func(cmd *cobra.Command, args []string) {
		channelFlag := cmd.Flag("channel")
		channel := channelFlag.Value.String()
		volume := getVolume(channel)
		fmt.Println(volume)
	},
}

var Channel string

func init() {
	audioCmd.AddCommand(volumeCmd)

	// Here you will define your flags and configuration settings.

}
