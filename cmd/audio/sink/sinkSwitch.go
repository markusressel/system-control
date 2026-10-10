package sink

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the default sink",
	Long: `Switches the default audio sink and moves all existing audio streams to the given one.
You can specify the audio sink using its index, but also using other strings that occur in its description:

> system-control audio sink switch "headphone"

> system-control audio sink switch "NVIDIA"`,
	Args: cobra.ExactArgs(1),
	RunE: commandutil.OpRunE("audio.sink.switch"),
}

func init() {
	SinkCmd.AddCommand(switchCmd)
}
