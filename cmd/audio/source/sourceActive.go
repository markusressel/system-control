package source

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var columns []string

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Show the default source",
	Long: `Show the source (input, f.e. a microphone) that is currently the default:

> system-control audio source active -c name
alsa_input.pci-0000_11_00.4.analog-stereo`,
	Args: cobra.NoArgs,
	RunE: commandutil.OpRunE("audio.source.active"),
}

func init() {
	activeCmd.Flags().StringSliceVarP(
		&columns,
		"columns", "c",
		[]string{"id", "name", "description"},
		"Columns to print (id,name,description)",
	)
	commandutil.AddWatchFlag(activeCmd)

	SourceCmd.AddCommand(activeCmd)
}
