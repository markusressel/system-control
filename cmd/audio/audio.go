package audio

import (
	"github.com/markusressel/system-control/cmd/audio/device"
	"github.com/markusressel/system-control/cmd/audio/sink"
	"github.com/markusressel/system-control/cmd/audio/volume"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:              "audio",
	Short:            "Control System Audio",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	Command.AddCommand(device.DeviceCmd)
	Command.AddCommand(sink.SinkCmd)
	Command.AddCommand(volume.VolumeCmd)
}
