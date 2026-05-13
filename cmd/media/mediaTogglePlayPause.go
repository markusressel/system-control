package media

import "github.com/spf13/cobra"

var togglePlayPauseCmd = &cobra.Command{
	Use:     "togglePlayPause",
	Aliases: []string{"toggle-play-pause"},
	Short:   "Toggle play/pause",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("play-pause", true)
	},
}

func init() {
	Command.AddCommand(togglePlayPauseCmd)
}
