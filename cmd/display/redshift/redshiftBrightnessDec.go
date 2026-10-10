package redshift

import (
	"fmt"

	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftBrightnessDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift brightness.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.brightness.dec"),
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

func validateStep(step float64) error {
	if step <= 0 {
		return fmt.Errorf("step size must be a positive number (found: %f)", step)
	}
	if step > 1.0 {
		return fmt.Errorf("step size cannot exceed 1.0 (found: %f)", step)
	}
	return nil
}
