package fan

import "github.com/spf13/cobra"

var Command = &cobra.Command{
	Use:   "fan",
	Short: "Control fan settings",
	Long:  ``,
}

func init() {
	Command.AddCommand(modeCmd)
}
