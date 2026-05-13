package media

import "github.com/spf13/cobra"

var positionCmd = &cobra.Command{
	Use:   "position",
	Short: "Show playback position",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("position", true)
	},
}

func init() {
	Command.AddCommand(positionCmd)
}
