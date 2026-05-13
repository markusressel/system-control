package media

import "github.com/spf13/cobra"

var previousCmd = &cobra.Command{
	Use:   "previous",
	Short: "Go to previous track",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("previous", true)
	},
}

func init() {
	Command.AddCommand(previousCmd)
}
