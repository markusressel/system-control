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

// toggleMuteCmd represents the toggleMute command
var toggleMuteCmd = &cobra.Command{
	Use:   "toggle-mute",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		channelFlag := cmd.Flag("channel")
		channel := channelFlag.Value.String()
		isMuted := isMuted(channel)
		setMuted(channel, !isMuted)
	},
}

func init() {
	audioCmd.AddCommand(toggleMuteCmd)

	// Here you will define your flags and configuration settings.

}
