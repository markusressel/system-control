package session

import (
	"fmt"

	"github.com/spf13/cobra"
)

var lockedCmd = &cobra.Command{
	Use:   "locked",
	Short: "Check whether the desktop session is locked",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		running, err := isProcessRunning("i3lock")
		if err != nil {
			return err
		}

		if running {
			fmt.Println("yes")
		} else {
			fmt.Println("no")
		}

		return nil
	},
}

func init() {
	Command.AddCommand(lockedCmd)
}
