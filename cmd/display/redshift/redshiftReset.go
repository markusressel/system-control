package redshift

import (
	internalredshift "github.com/markusressel/system-control/internal/redshift"
	"github.com/spf13/cobra"

	"github.com/markusressel/system-control/internal/daemon"
)

var redshiftResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the currently applied redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		unlock, err := daemon.AcquireLock(internalredshift.LockName)
		if err != nil {
			return err
		}
		defer unlock()

		displays, err := internalredshift.ParseDisplays(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			err := internalredshift.Reset(display)
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
