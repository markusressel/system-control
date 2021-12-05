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
package audio

import (
	"github.com/markusressel/system-control/internal"
	"github.com/spf13/cobra"
	"strconv"
)

// decVolumeCmd represents the dec command
var decVolumeCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrement audio volume",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cardFlag := cmd.Flag("card")
		card := cardFlag.Value.String()
		cardInt, _ := strconv.Atoi(card)

		channelFlag := cmd.Flag("channel")
		channel := channelFlag.Value.String()

		volume := internal.GetVolume(cardInt, channel)
		change := internal.CalculateAppropriateVolumeChange(volume, false)
		internal.SetVolume(cardInt, channel, volume-change)
	},
}

func init() {
	volumeCmd.AddCommand(decVolumeCmd)

	// Here you will define your flags and configuration settings.

}
