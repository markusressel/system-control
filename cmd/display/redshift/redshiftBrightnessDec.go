package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var redshiftBrightnessDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift brightness.",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := stepInputValidator(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			lastSetBrightness := getLastSetBrightness(display)

			rawNew := lastSetBrightness - stepFloat
			newBrightness := clampBrightnessToConfig(rawNew)
			err = ApplyRedshift(display, -1, newBrightness, -1)
			if err != nil {
				return err
			}

			fmt.Println(newBrightness)
		}

		return nil
	},
}

func init() {
	redshiftBrightnessDecCmd.PersistentFlags().Float64VarP(
		&stepFloat,
		"step", "s",
		0.1,
		"Step size to increase the brightness by (between 0.1 and 1.0)",
	)

	brightnessCmd.AddCommand(redshiftBrightnessDecCmd)
}

func stepInputValidator(cmd *cobra.Command, args []string) error {
	if stepFloat <= 0 {
		return fmt.Errorf("step size must be a positive number (found: %f)", stepFloat)
	}
	if stepFloat > 1.0 {
		return fmt.Errorf("step size cannot exceed 1.0 (found: %f)", stepFloat)
	}
	return nil
}
