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
	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/audio/pulseaudio"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/spf13/cobra"
	"strconv"
)

type audioState struct {
	OutputType string
	Card       int
	Channel    string
	Volume     int
	Muted      bool
}

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current state of the given audio channel",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cardFlag := cmd.Flag("card")
		card := cardFlag.Value.String()
		cardInt, _ := strconv.Atoi(card)

		channelFlag := cmd.Flag("channel")
		channel := channelFlag.Value.String()

		currentVolume := pulseaudio.GetVolume(cardInt, channel)
		muted := pulseaudio.IsMuted(cardInt, channel)

		key := computeKey(audio.IsHeadphoneConnected(), card, channel)
		data := audioState{
			OutputType: "OutputType",
			Card:       cardInt,
			Channel:    channel,
			Volume:     currentVolume,
			Muted:      muted,
		}
		err := persistence.SaveStruct(key, &data)

		return err
	},
}

func init() {
	VolumeCmd.AddCommand(saveCmd)
}
