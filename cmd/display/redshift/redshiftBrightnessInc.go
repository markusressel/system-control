package redshift

import (
	"fmt"

	"github.com/markusressel/system-control/internal/commandutil"
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
	RunE: commandutil.LockedRunE("cmd-display-redshift", commandutil.MutatorRunE(func(cmd *cobra.Command, args []string) ([]commandutil.MutatorResult, error) {
		displays, err := parseDisplayParam(display)
		if err != nil {
			return nil, err
		}

		var results []commandutil.MutatorResult
		for _, display := range displays {
			lastSetBrightness := getLastSetBrightness(display)

			rawNew := lastSetBrightness + stepFloat
			newBrightness := clampBrightnessToConfig(rawNew)
			err = ApplyRedshift(display, -1, newBrightness, -1)
			if err != nil {
				return nil, err
			}

			results = append(results, commandutil.MutatorResult{
				CommandId: "display.redshift.brightness",
				Flags:     map[string]string{"display": display.Name},
				Value:     fmt.Sprintf("%.2f", newBrightness),
			})
		}

		return results, nil
	})),
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
