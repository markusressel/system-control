package source

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch the default source",
	Long: `Makes the source (input, f.e. a microphone) whose name or description contains the given text the default,
and moves all existing recording streams to it. They keep following the default, like with "audio sink switch":

> system-control audio source switch "alsa_input.pci-0000_11_00.4.analog-stereo"`,
	Args: cobra.ExactArgs(1),
	RunE: commandutil.OpRunE("audio.source.switch"),
}

func init() {
	SourceCmd.AddCommand(switchCmd)
}
