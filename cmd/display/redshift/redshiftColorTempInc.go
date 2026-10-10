package redshift

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/markusressel/system-control/internal/daemon"
)

var redshiftColorTempIncCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the currently applied redshift color temperature.",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return colorTempStepInputValidator(cmd, args)
	},
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
			lastSetColorTemperature := getLastSetColorTemperature(display)

			rawNew := lastSetColorTemperature + stepInt
			newColorTemperature := clampColorTemperatureToConfig(rawNew)
			err = ApplyRedshift(display, newColorTemperature, -1, -1)
			if err != nil {
				return err
			}

			fmt.Println(newColorTemperature)
		}

		return nil
	},
}

func init() {
	redshiftColorTempIncCmd.PersistentFlags().Int64VarP(
		&stepInt,
		"step", "s",
		500,
		"Step size to increase the color temperature by",
	)

	colorTempCmd.AddCommand(redshiftColorTempIncCmd)
}
