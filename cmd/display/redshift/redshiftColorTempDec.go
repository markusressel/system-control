package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var redshiftColorTempDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift color temperature.",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return colorTempStepInputValidator(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			lastSetColorTemperature := getLastSetColorTemperature(display)

			rawNew := lastSetColorTemperature - stepInt
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
	redshiftColorTempDecCmd.PersistentFlags().Int64VarP(
		&stepInt,
		"step", "s",
		500,
		"Step size to decrease the color temperature by",
	)

	colorTempCmd.AddCommand(redshiftColorTempDecCmd)
}

func colorTempStepInputValidator(cmd *cobra.Command, args []string) error {
	if stepInt <= 0 {
		return fmt.Errorf("step size must be a positive number (found: %d)", stepInt)
	}
	return nil
}
