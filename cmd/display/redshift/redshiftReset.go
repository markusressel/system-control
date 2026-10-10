package redshift

import (
	"github.com/spf13/cobra"

	"github.com/markusressel/system-control/internal/daemon"
)

var redshiftResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the currently applied redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		unlock, err := daemon.AcquireLock(redshiftLockName)
		if err != nil {
			return err
		}
		defer unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			err := ResetRedshift(display)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(redshiftResetCmd)
}
