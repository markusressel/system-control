package redshift

import (
	"github.com/spf13/cobra"
)

var redshiftResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the currently applied redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

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
