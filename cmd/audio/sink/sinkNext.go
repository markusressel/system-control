package sink

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to the next sink after the currently active one",
	Long: `Switches the default audio sink and moves all existing audio streams to the next available one.

> system-control audio sink next`,
	Args: cobra.NoArgs,
	// TODO: add "skip" string glob parameter, so that the user can skip certain sinks when switching to the next one
	RunE: commandutil.OpRunE("audio.sink.next"),
}

func init() {
	SinkCmd.AddCommand(nextCmd)
}
