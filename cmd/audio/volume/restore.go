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
	"strconv"

	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/audio/pulseaudio"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the state of the given audio channel from a previous save",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cardFlag := cmd.Flag("card")
		card := cardFlag.Value.String()
		cardInt, _ := strconv.Atoi(card)

		channelFlag := cmd.Flag("channel")
		channel := channelFlag.Value.String()

		headphonesConnected := audio.IsHeadphoneConnected()
		key := computeKey(headphonesConnected, card, channel)

		data := audioState{}
		err := persistence.ReadStruct(key, &data)
		if err != nil {
			return err
		}

		err = pulseaudio.SetMuted(cardInt, channel, data.Muted)
		err = pulseaudio.SetVolume(cardInt, channel, data.Volume)

		return err
	},
}

func computeKey(headphonesConnected bool, card string, channel string) string {
	var speakerType string
	if headphonesConnected {
		speakerType = "headphones"
	} else {
		speakerType = "speaker"
	}

	return speakerType + "_" + card + "_" + channel
}

func init() {
	VolumeCmd.AddCommand(restoreCmd)
}
