package volume

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var device string
var stream string

var VolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Show the current volume",
	RunE: func(cmd *cobra.Command, args []string) error {
		state := pipewire.PwDump()

		volume, err := state.GetVolumeByName(device)
		if err != nil {
			return err
		}
		volume = util.RoundToTwoDecimals(volume)
		volumeAsInt := (int)(volume * 100)
		fmt.Println(volumeAsInt)
		return nil
	},
}

func init() {
	VolumeCmd.PersistentFlags().StringVarP(
		&device,
		"device", "d",
		"",
		"Device Name/Description",
	)

	VolumeCmd.PersistentFlags().StringVarP(
		&stream,
		"stream", "s",
		"",
		"Stream Name/Description",
	)
}
