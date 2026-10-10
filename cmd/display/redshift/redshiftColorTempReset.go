package redshift

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/markusressel/system-control/internal/daemon"
)

var redshiftColorTempResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the display color temperature to 6500K.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
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
			err = ApplyRedshift(display, 6500, -1, -1)
			if err != nil {
				return err
			}

			fmt.Println(6500)
		}

		return nil
	},
}

func init() {
	colorTempCmd.AddCommand(redshiftColorTempResetCmd)
}
