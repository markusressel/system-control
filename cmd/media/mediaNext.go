package media

import "github.com/spf13/cobra"

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to next track",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("next", true)
	},
}

func init() {
	Command.AddCommand(nextCmd)
}
