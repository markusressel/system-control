package sink

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var previousCmd = &cobra.Command{
	Use:   "previous",
	Short: "Switch to the previous sink before the currently active one",
	Long: `Switches the default audio sink and moves all existing audio streams to the previous available one.

> system-control audio sink previous`,
	Args: cobra.NoArgs,
	// TODO: add "skip" string glob parameter, so that the user can skip certain sinks when switching to the next one
	RunE: commandutil.OpRunE("audio.sink.previous"),
}

func init() {
	SinkCmd.AddCommand(previousCmd)
}
