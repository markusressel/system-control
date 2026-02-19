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
