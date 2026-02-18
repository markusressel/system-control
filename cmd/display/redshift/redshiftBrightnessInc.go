/*
 * system-control
 * Copyright (c) 2019. Markus Ressel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package redshift

import (
	"errors"
	"fmt"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var redshiftBrightnessIncCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the currently applied redshift brightness.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		configPath := configuration.DetectAndReadConfigFile()
		//ui.Info("Using configuration file at: %s", configPath)
		config := configuration.LoadConfig()
		err = configuration.Validate(configPath)
		if err != nil {
			//ui.FatalWithoutStacktrace(err.Error())
		}

		redshiftConfig, err := util.ReadRedshiftConfig()
		if err != nil {
			return err
		}

		colorTemperature := CalculateTargetColorTemperature(
			config.Redshift,
			redshiftConfig,
		)

		if colorTemperature != -1 && (colorTemperature < 1000 || colorTemperature > 25000) {
			return errors.New("color temperature must be between 1000 and 25000")
		}

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			err = ApplyRedshift(display, colorTemperature, -1, -1)
			if err != nil {
				return err
			}

			// print current values, if any of them were changed
			lastSetColorTemperature := getLastSetColorTemperature(display)
			lastSetBrightness := getLastSetBrightness(display)
			lastSetGamma := getLastSetGamma(display)

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
	Command.AddCommand(redshiftBrightnessIncCmd)
}
