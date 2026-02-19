package sink

import (
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var previousCmd = &cobra.Command{
	Use:   "previous",
	Short: "Switch to the previous sink before the currently active one",
	Long: `Switches the default audio sink and moves all existing audio streams to the previous available one.

> system-control audio sink previous`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipewire.RotateActiveSinkPipewire(true)
	},
}

func init() {
	SinkCmd.AddCommand(previousCmd)
}
