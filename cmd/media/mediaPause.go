package media

import "github.com/spf13/cobra"

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause media",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("pause", true)
	},
}

func init() {
	Command.AddCommand(pauseCmd)
}
