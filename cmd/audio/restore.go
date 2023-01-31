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
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"strconv"
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

		headphonesConnected := util.IsHeadphoneConnected()
		key := computeKey(headphonesConnected, Card, Channel)

		data := audioState{}
		err := persistence.ReadStruct(key, &data)
		if err != nil {
			return err
		}

		err = util.SetMuted(cardInt, channel, data.Muted)
		err = util.SetVolume(cardInt, Channel, data.Volume)

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
	Command.AddCommand(restoreCmd)
}
