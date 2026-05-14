package session

import (
	"github.com/spf13/cobra"
)

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock the current desktop session",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terminateProcessesByName("i3lock")
	},
}

func init() {
	Command.AddCommand(unlockCmd)
}
