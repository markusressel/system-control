package session

import (
	internalsession "github.com/markusressel/system-control/internal/session"
	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock the current desktop session",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		isDetachedRunner, err := cmd.Flags().GetBool(internalsession.LockRunnerFlag)
		if err != nil {
			return err
		}
		useDetached, err := cmd.Flags().GetBool(internalsession.LockDetachedFlag)
		if err != nil {
			return err
		}

		if useDetached && !isDetachedRunner {
			return internalsession.StartDetachedLockProcess()
		}

		return currentSession().Lock()
	},
}

func init() {
	Command.AddCommand(lockCmd)
	lockCmd.Flags().Bool(internalsession.LockDetachedFlag, false, "run lock command in a detached background process")
	lockCmd.Flags().Bool(internalsession.LockRunnerFlag, false, "internal flag used to run detached lock process")
	_ = lockCmd.Flags().MarkHidden(internalsession.LockRunnerFlag)
}
