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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/nathan-osman/go-sunrise"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var redshiftUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the currently applied redshift based on the current time of day.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		configPath := configuration.DetectAndReadConfigFile()
		//ui.Info("Using configuration file at: %s", configPath)
		config := configuration.LoadConfig()
		err = configuration.Validate(configPath)
		if err != nil {
			//ui.FatalWithoutStacktrace(err.Error())
		}

		redshiftConfig, err := readRedshiftConfig()
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

			// print current values
			//fmt.Printf("Color Temperature: %d -> %d\n", lastSetColorTemperature, colorTemperature)
			//fmt.Printf("Brightness: %.2f -> %.2f\n", lastSetBrightness, brightness)
			//fmt.Printf("Gamma: %.2f -> %.2f\n", lastSetGamma, gamma)
		}

		return nil
	},
}

type RedshiftConfigBlock struct {
	DayColorTemperature   int64  `toml:"temp-day"`
	NightColorTemperature int64  `toml:"temp-night"`
	LocationProvider      string `toml:"location-provider"`
}

type RedshiftManualConfigBlock struct {
	Lat float64 `toml:"lat"`
	Lon float64 `toml:"lon"`
}

type RedshiftConfig struct {
	Redshift RedshiftConfigBlock       `toml:"redshift"`
	Manual   RedshiftManualConfigBlock `toml:"manual"`
}

func readRedshiftConfig() (RedshiftConfig, error) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config/redshift.conf")
	doc, err := os.ReadFile(configPath)

	lines := strings.Split(string(doc), "\n")
	var linesWithoutComments []string
	for _, line := range lines {
		// work around for toml parser not supporting comments starting with ;
		if strings.HasPrefix(line, ";") {
			continue
		}

		// work around for toml parser not supporting unquoted values
		if strings.Contains(line, "=manual") {
			line = strings.ReplaceAll(line, "=manual", "='manual'")
		}
		if strings.Contains(line, "=randr") {
			line = strings.ReplaceAll(line, "=randr", "='randr'")
		}

		linesWithoutComments = append(linesWithoutComments, line)
	}

	configWithoutComments := strings.Join(linesWithoutComments, "\n")

	var cfg RedshiftConfig
	err = toml.Unmarshal([]byte(configWithoutComments), &cfg)
	return cfg, err
}

const (
	// TransitionElevationThreshold is the sun's elevation threshold at which the color temperature should be fully transitioned.
	TransitionElevationThreshold = 20
)

func CalculateTargetColorTemperature(
	_redshiftConfig configuration.RedshiftConfig,
	redshiftConfig RedshiftConfig,
) int64 {
	elevation := sunrise.Elevation(
		redshiftConfig.Manual.Lat,
		redshiftConfig.Manual.Lon,
		time.Now(),
	)

	targetColor := redshiftConfig.Redshift.DayColorTemperature
	if elevation < 0 {
		targetColor = redshiftConfig.Redshift.NightColorTemperature
	} else if elevation > TransitionElevationThreshold {
		targetColor = redshiftConfig.Redshift.DayColorTemperature
	} else {
		targetColor = redshiftConfig.Redshift.NightColorTemperature + int64((float64(redshiftConfig.Redshift.DayColorTemperature)-float64(redshiftConfig.Redshift.NightColorTemperature))*(elevation/TransitionElevationThreshold))
	}

	return targetColor
}

func init() {
	redshiftCmd.AddCommand(redshiftUpdateCmd)
}
