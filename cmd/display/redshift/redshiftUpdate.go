package redshift

import (
	"errors"
	"fmt"
	internalredshift "github.com/markusressel/system-control/internal/redshift"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var redshiftUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the currently applied redshift based on the current time of day.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		unlock, err := daemon.AcquireLock(internalredshift.LockName)
		if err != nil {
			return err
		}
		defer unlock()

		config := configuration.Get()

		redshiftConfig, err := util.ReadRedshiftConfig()
		if err != nil {
			return err
		}

		colorTemperature := internalredshift.CalculateTargetColorTemperature(
			config.Redshift,
			redshiftConfig,
		)

		if colorTemperature != -1 && (colorTemperature < internalredshift.MinColorTemperature || colorTemperature > internalredshift.MaxColorTemperature) {
			return errors.New("color temperature must be between 1000 and 25000")
		}

		displays, err := internalredshift.ParseDisplays(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			err = internalredshift.Apply(display, colorTemperature, -1, -1)
			if err != nil {
				return err
			}

			// print current values, if any of them were changed
			lastSetColorTemperature := internalredshift.LastColorTemperature(display)
			lastSetBrightness := internalredshift.LastBrightness(display)
			lastSetGamma := internalredshift.LastGamma(display)

			if colorTemperature != -1 && colorTemperature != lastSetColorTemperature {
				fmt.Printf("Color Temperature: %d -> %d\n", lastSetColorTemperature, colorTemperature)
				fmt.Printf("Brightness: %.2f\n", lastSetBrightness)
				fmt.Printf("Gamma: %.2f\n", lastSetGamma)
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(redshiftUpdateCmd)
}
