package media

import "github.com/spf13/cobra"

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show media status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("status", true)
	},
}

func init() {
	Command.AddCommand(statusCmd)
}
