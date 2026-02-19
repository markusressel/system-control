package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var redshiftBrightnessResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the display brightness to 1.0.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			err = ApplyRedshift(display, -1, 1.0, -1)
			if err != nil {
				return err
			}

			fmt.Println(1.0)
		}

		return nil
	},
}

func init() {
	brightnessCmd.AddCommand(redshiftBrightnessResetCmd)
}
