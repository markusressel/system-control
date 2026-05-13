package media

import "github.com/spf13/cobra"

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play media",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("play", true)
	},
}

func init() {
	Command.AddCommand(playCmd)
}
