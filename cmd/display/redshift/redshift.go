package redshift

import (
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"

	"github.com/gofrs/flock"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

const (
	KeyRedshiftColorTemp  = "redshift.colorTemperature"
	KeyRedshiftBrightness = "redshift.brightness"
	KeyRedshiftGamma      = "redshift.gamma"
)

var (
	display string

	colorTemperature int64
	brightness       float64
	gamma            float64

	brightnessValue float64
	stepFloat             = 0.1
	stepInt         int64 = 500

	redshiftLock *flock.Flock
)

var Command = &cobra.Command{
	Use:   "redshift",
	Short: "Apply the given redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		if colorTemperature != -1 && (colorTemperature < 1000 || colorTemperature > 25000) {
			return errors.New("color temperature must be between 1000 and 25000")
		}

		if brightness != -1 && (brightness < 0.1 || brightness > 1.0) {
			return errors.New("brightness must be between 0.1 and 1.0")
		}

		if gamma != -1 && (gamma < 0.1 || gamma > 2.0) {
			return errors.New("gamma must be between 0.1 and 2.0")
		}

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {

			lastSetColorTemperature := getLastSetColorTemperature(display)
			lastSetBrightness := getLastSetBrightness(display)
			lastSetGamma := getLastSetGamma(display)

			err := ApplyRedshift(display, colorTemperature, brightness, gamma)
			if err != nil {
				return err
			}

			// print current values
			fmt.Printf("Display: %s\n", display)
			if colorTemperature != -1 {
				fmt.Printf("  Color Temperature: %d -> %d\n", lastSetColorTemperature, colorTemperature)
			} else {
				fmt.Printf("  Color Temperature: %d\n", lastSetColorTemperature)
			}
			if brightness != -1 {
				fmt.Printf("  Brightness: %.2f -> %.2f\n", lastSetBrightness, brightness)
			} else {
				fmt.Printf("  Brightness: %.2f\n", lastSetBrightness)
			}
			if gamma != -1 {
				fmt.Printf("  Gamma: %.2f -> %.2f\n", lastSetGamma, gamma)
			} else {
				fmt.Printf("  Gamma: %.2f\n", lastSetGamma)
			}
		}

		return nil
	},
}

func parseDisplayParam(display string) (result []util.DisplayInfo, err error) {
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

func getLastSetColorTemperature(display util.DisplayInfo) int64 {
	key := KeyRedshiftColorTemp + "." + display.Name
	lastSetColorTemperature, err := persistence.ReadInt(key)
	if err != nil {
		lastSetColorTemperature = -1
	}
	return lastSetColorTemperature
}

func getLastSetBrightness(display util.DisplayInfo) float64 {
	key := KeyRedshiftBrightness + "." + display.Name
	lastSetBrightness, err := persistence.ReadFloat(key)
	if err != nil {
		lastSetBrightness = -1
	}
	if lastSetBrightness < 0.1 {
		lastSetBrightness = 0.1
		saveLastSetBrightness(display, lastSetBrightness)
	}
	if lastSetBrightness > 1.0 {
		lastSetBrightness = 1.0
		saveLastSetBrightness(display, lastSetBrightness)
	}
	return lastSetBrightness
}

func getLastSetGamma(display util.DisplayInfo) float64 {
	key := KeyRedshiftGamma + "." + display.Name
	lastSetGamma, err := persistence.ReadFloat(key)
	if err != nil {
		lastSetGamma = 1.0
	}
	if lastSetGamma < 0.1 {
		lastSetGamma = 0.1
		saveLastSetGamma(display, lastSetGamma)
	}
	if lastSetGamma > 2.0 {
		lastSetGamma = 2.0
		saveLastSetGamma(display, lastSetGamma)
	}
	return lastSetGamma
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

func ApplyRedshift(display util.DisplayInfo, colorTemperature int64, brightness float64, gamma float64) (err error) {
	if colorTemperature == -1 {
		colorTemperature = getLastSetColorTemperature(display)
	}
	if brightness == -1 {
		brightness = getLastSetBrightness(display)
	}
	if gamma == -1 {
		gamma = getLastSetGamma(display)
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

func ResetRedshift(display util.DisplayInfo) (err error) {
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

var (
	TEMP_PATH = os.TempDir() + "/system-control"
)

func lockRedshift() *flock.Flock {
	err := os.MkdirAll(TEMP_PATH, 0755)
	if err != nil {
		log.Fatalf("Could not create temp directory: %v", err)
	}
	fileLock := flock.New(TEMP_PATH + "/cmd-display-redshift.lock")
	err = fileLock.Lock()
	if err != nil {
		log.Fatalf("Could not acquire lock: %v", err)
	}

	return fileLock
}

func init() {
	Command.PersistentFlags().StringVarP(
		&display,
		"display", "d",
		"",
		"Display",
	)

	Command.PersistentFlags().Int64VarP(
		&colorTemperature,
		"temperature", "t",
		-1,
		"Color Temperature",
	)

	Command.PersistentFlags().Float64VarP(
		&brightness,
		"brightness", "b",
		-1,
		"Brightness",
	)

	Command.PersistentFlags().Float64VarP(
		&gamma,
		"gamma", "g",
		-1,
		"Gamma",
	)
}
