package session

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

const lockScreenScript = "/home/markus/.custom/bin/lock-screen"

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock the current desktop session",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.ExecCommandAndFork(lockScreenScript)
	},
}

func init() {
	Command.AddCommand(lockCmd)
}
