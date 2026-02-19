package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var redshiftGammaDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift gamma.",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return gammaStepInputValidator(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			lastSetGamma := getLastSetGamma(display)

			rawNew := lastSetGamma - stepFloat
			newGamma := clampGammaToConfig(rawNew)
			err = ApplyRedshift(display, -1, -1, newGamma)
			if err != nil {
				return err
			}

			fmt.Println(newGamma)
		}

		return nil
	},
}

func init() {
	redshiftGammaDecCmd.PersistentFlags().Float64VarP(
		&stepFloat,
		"step", "s",
		0.1,
		"Step size to decrease the gamma by",
	)

	gammaCmd.AddCommand(redshiftGammaDecCmd)
}

func gammaStepInputValidator(cmd *cobra.Command, args []string) error {
	if stepFloat <= 0 {
		return fmt.Errorf("step size must be a positive number (found: %f)", stepFloat)
	}
	return nil
}
