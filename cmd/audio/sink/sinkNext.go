package sink

import (
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to the next sink after the currently active one",
	Long: `Switches the default audio sink and moves all existing audio streams to the next available one.

> system-control audio sink next`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipewire.RotateActiveSinkPipewire(false)
	},
}

func init() {
	SinkCmd.AddCommand(nextCmd)
}
