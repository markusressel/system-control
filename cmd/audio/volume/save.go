package volume

import (
	"strconv"

	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/audio/pulseaudio"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/spf13/cobra"
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
