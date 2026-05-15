package session

import (
	"fmt"

	"github.com/spf13/cobra"
)

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock the current desktop session",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		terminateErr := terminateProcessesByName("i3lock")
		restoreErr := restoreSessionDPMSTimeout()

		if terminateErr != nil && restoreErr != nil {
			return fmt.Errorf("failed to unlock session: %w; additionally failed to restore session DPMS timeout: %v", terminateErr, restoreErr)
		}

		if terminateErr != nil {
			return terminateErr
		}

		return restoreErr
	},
}

func init() {
	Command.AddCommand(unlockCmd)
}
