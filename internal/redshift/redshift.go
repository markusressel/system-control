// Package redshift sets the color temperature, brightness and gamma of displays with the
// `redshift` tool and remembers the last values per display.
package redshift

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
	"github.com/nathan-osman/go-sunrise"
)

const (
	KeyRedshiftColorTemp  = "redshift.colorTemperature"
	KeyRedshiftBrightness = "redshift.brightness"
	KeyRedshiftGamma      = "redshift.gamma"
)

// LockName serializes everything that applies a redshift, across processes.
const LockName = "cmd-display-redshift"

// minimum and maximum color temperature in Kelvin accepted when setting a value
const (
	MinColorTemperature = 1000
	MaxColorTemperature = 25000
)

// TransitionElevationThreshold is the sun's elevation threshold at which the color temperature should be fully transitioned.
const TransitionElevationThreshold = 20

// ParseDisplays returns the display of the given name, or all displays if the name is empty.
func ParseDisplays(display string) (result []util.DisplayInfo, err error) {
	if len(display) > 0 {
		foundDisplayName, err := findDisplay(display)
		if err != nil {
			return nil, err
		}
		if foundDisplayName != nil {
			result = []util.DisplayInfo{*foundDisplayName}
		}
	} else {
		return util.GetDisplays()
	}

	return result, nil
}

// findDisplay finds a display by name
func findDisplay(d string) (*util.DisplayInfo, error) {
	knownDisplays, err := util.GetDisplays()
	if err != nil {
		return nil, err
	}
	for _, knownDisplay := range knownDisplays {
		if knownDisplay.Name == d {
			return &knownDisplay, nil
		}
	}
	return nil, fmt.Errorf("display named %s not found", d)
}

// LastColorTemperature returns the color temperature last applied to the display, -1 if none.
func LastColorTemperature(display util.DisplayInfo) int64 {
	key := KeyRedshiftColorTemp + "." + display.Name
	lastSetColorTemperature, err := persistence.ReadInt(key)
	if err != nil {
		lastSetColorTemperature = -1
	}
	return lastSetColorTemperature
}

// LastBrightness returns the brightness last applied to the display.
func LastBrightness(display util.DisplayInfo) float64 {
	key := KeyRedshiftBrightness + "." + display.Name
	lastSetBrightness, err := persistence.ReadFloat(key)
	if err != nil {
		lastSetBrightness = -1
	}
	// Only clamp, never write back: a failed read must not overwrite the stored value.
	return min(max(lastSetBrightness, 0.1), 1.0)
}

// LastGamma returns the gamma last applied to the display.
func LastGamma(display util.DisplayInfo) float64 {
	key := KeyRedshiftGamma + "." + display.Name
	lastSetGamma, err := persistence.ReadFloat(key)
	if err != nil {
		lastSetGamma = 1.0
	}
	return min(max(lastSetGamma, 0.1), 2.0)
}

func saveLastSetColorTemperature(display util.DisplayInfo, colorTemperature int64) error {
	key := KeyRedshiftColorTemp + "." + display.Name
	return persistence.SaveInt(key, int(colorTemperature))
}

func saveLastSetBrightness(display util.DisplayInfo, brightness float64) error {
	key := KeyRedshiftBrightness + "." + display.Name
	return persistence.SaveFloat(key, brightness)
}

func saveLastSetGamma(display util.DisplayInfo, gamma float64) error {
	key := KeyRedshiftGamma + "." + display.Name
	return persistence.SaveFloat(key, gamma)
}

// ClampBrightnessToConfig limits a brightness to the configured range.
func ClampBrightnessToConfig(value float64) float64 {
	config := configuration.Get()
	minBr := config.Redshift.Brightness.MinimumBrightness
	maxBr := config.Redshift.Brightness.MaximumBrightness
	newBrightnessValue := util.Clamp(value, minBr, maxBr)
	newBrightnessValue = util.RoundTo2DP(newBrightnessValue)
	return newBrightnessValue
}

// ClampGammaToConfig limits a gamma to the configured range.
func ClampGammaToConfig(value float64) float64 {
	config := configuration.Get()
	minG := config.Redshift.Gamma.MinimumGamma
	maxG := config.Redshift.Gamma.MaximumGamma
	newGammaValue := util.Clamp(value, minG, maxG)
	newGammaValue = util.RoundTo2DP(newGammaValue)
	return newGammaValue
}

// ClampColorTemperatureToConfig limits a color temperature to the configured range.
func ClampColorTemperatureToConfig(colorTemperature int64) int64 {
	config := configuration.Get()
	minCT := config.Redshift.ColorTemperature.MinimumColorTemperature
	maxCT := config.Redshift.ColorTemperature.MaximumColorTemperature
	return util.Clamp(colorTemperature, minCT, maxCT)
}

// Apply sets color temperature, brightness and gamma of the display (-1 keeps the last value)
// and remembers them.
func Apply(display util.DisplayInfo, colorTemperature int64, brightness float64, gamma float64) (err error) {
	if colorTemperature == -1 {
		colorTemperature = LastColorTemperature(display)
	}
	if brightness == -1 {
		brightness = LastBrightness(display)
	}
	if gamma == -1 {
		gamma = LastGamma(display)
	}

	displays, err := util.GetDisplays()
	if err != nil {
		return err
	}
	displayIndex := slices.IndexFunc(displays, func(d util.DisplayInfo) bool { return d.Name == display.Name })

	err = SetRedshiftCBG(displayIndex, colorTemperature, brightness, gamma)
	if err != nil {
		return err
	}

	err = saveLastSetColorTemperature(display, colorTemperature)
	if err != nil {
		return err
	}
	err = saveLastSetBrightness(display, brightness)
	if err != nil {
		return err
	}
	err = saveLastSetGamma(display, gamma)
	if err != nil {
		return err
	}

	// a running daemon has to learn about the persisted display state
	daemon.Notify("display.")
	return nil
}

// SetRedshiftCBG sets the redshift color temperature, brightness and gamma
// colorTemperature: the color temperature in Kelvin, between 1000 and 25000 (-1 to ignore, 6500 is default
// brightness: the brightness value between 0.1 and 1.0 (-1 to ignore, 1.0 is default)
// gamma: the gamma value between 0.1 and 1.0 (-1 to ignore, 1.0 is default)
// immediate: apply the changes immediately, without transition
func SetRedshiftCBG(displayIndex int, colorTemperature int64, brightness float64, gamma float64) error {
	args := []string{
		"-x", // reset previous "mode"
		"-P", // reset previous gamma ramps
		"-o", // one shot mode
	}

	if displayIndex > -1 {
		// -m randr:crtc=1
		args = append(args, "-m", fmt.Sprintf("randr:crtc=%d", displayIndex))
	}

	if colorTemperature != -1 {
		// set color temperature
		args = append(args, "-O", fmt.Sprintf("%d", colorTemperature))
	}

	if brightness != -1 {
		// set brightness
		args = append(args, "-b", strconv.FormatFloat(brightness, 'f', -1, 64))
	}

	if gamma != -1 {
		// set gamma
		args = append(args, "-g", strconv.FormatFloat(gamma, 'f', -1, 64))
	}

	if len(args) == 0 {
		return errors.New("no changes to apply")
	}
	_, err := util.ExecCommand("redshift", args...)
	return err
}

// Reset removes any redshift from the display.
func Reset(display util.DisplayInfo) (err error) {
	args := []string{
		"-x", // reset previous "mode"
		"-P", // reset previous gamma ramps
		"-r", // apply immediately
	}

	displays, err := util.GetDisplays()
	if err != nil {
		return err
	}
	displayIndex := slices.IndexFunc(displays, func(d util.DisplayInfo) bool { return d.Name == display.Name })

	if displayIndex > -1 {
		// -m randr:crtc=1
		args = append(args, "-m", fmt.Sprintf("randr:crtc=%d", displayIndex))
	}

	_, err = util.ExecCommand("redshift", args...)
	return err
}

// CalculateTargetColorTemperature returns the color temperature for the current elevation of
// the sun, between the night and the day color temperature of the redshift configuration.
func CalculateTargetColorTemperature(
	_redshiftConfig configuration.RedshiftConfig,
	redshiftConfig util.RedshiftConfig,
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
