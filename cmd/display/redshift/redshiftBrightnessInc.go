package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var redshiftBrightnessIncCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the currently applied redshift brightness.",
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

			rawNew := lastSetBrightness + stepFloat
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
	redshiftBrightnessIncCmd.PersistentFlags().Float64VarP(
		&stepFloat,
		"step", "s",
		0.1,
		"Step size to increase the brightness by (between 0.1 and 1.0)",
	)

	brightnessCmd.AddCommand(redshiftBrightnessIncCmd)
}
