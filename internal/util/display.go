package util

import (
	"errors"
	"log"
	"os"
	"strings"
)

func GetDisplays() (displays []string, err error) {
	result, err := ExecCommand(
		"xrandr",
		"--listmonitors",
	)
	if err != nil {
		return nil, err
	}

	resultLines := strings.Split(result, "\n")

	for i, line := range resultLines {
		if i == 0 {
			continue
		} else {
			segments := strings.Split(line, " ")
			displayName := segments[len(segments)-1]
			displays = append(displays, displayName)
		}
	}
	return displays, nil
}

func GetMaxBrightness() (int, error) {
	backlightName, err := findBacklight()
	if err != nil {
		return -1, err
	}
	maxBrightnessPath := computeBacklightPropertyPath(backlightName, MaxBrightness)
	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		return -1, err
	}
	return int(maxBrightness), nil
}

func GetBrightness() (int, error) {
	backlightName, err := findBacklight()
	if err != nil {
		return -1, err
	}
	brightnessPath := computeBacklightPropertyPath(backlightName, Brightness)
	brightness, err := ReadIntFromFile(brightnessPath)
	if err != nil {
		return -1, err
	}
	return int(brightness), nil
}

// SetBrightness sets the brightness of the main display to the given percentage
// Note: This function only works if the display has a max_brightness value
func SetBrightness(percentage int) error {
	backlightName, err := findBacklight()
	if err != nil {
		return err
	}
	maxBrightnessPath := computeBacklightPropertyPath(backlightName, MaxBrightness)
	brightnessPath := computeBacklightPropertyPath(backlightName, Brightness)

	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		return err
	}

	targetValue := int((float32(percentage) / 100.0) * float32(maxBrightness))
	return WriteIntToFile(targetValue, brightnessPath)
}

// setBrightnessRaw sets the brightness of the main display to the given value
func setBrightnessRaw(backlight string, brightness int) error {
	maxBrightness, err := GetMaxBrightness()
	if err != nil {
		return err
	}
	targetBrightness := brightness
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	brightnessPath := computeBacklightPropertyPath(backlight, Brightness)

	return WriteIntToFile(targetBrightness, brightnessPath)
}

func computeBacklightPropertyPath(backlight string, property string) string {
	return DisplayBacklightPath + string(os.PathSeparator) + backlight + string(os.PathSeparator) + property
}

// AdjustBrightness adjusts the brightness of the main display
func AdjustBrightness(change int) error {
	backlight, err := findBacklight()
	if err != nil {
		return err
	}

	maxBrightness, err := GetMaxBrightness()
	if err != nil {
		return err
	}
	currentBrightness, err := GetBrightness()
	if err != nil {
		return err
	}

	targetBrightness := currentBrightness + change
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	return setBrightnessRaw(backlight, targetBrightness)
}

func findBacklight() (string, error) {
	files, err := os.ReadDir(DisplayBacklightPath)
	if err != nil {
		return "", err
	}

	var backlightName string
	if len(files) == 0 {
		return "", errors.New("no backlight found")
	} else if len(files) == 1 {
		backlightName = files[0].Name()
	} else {
		// TODO: select first? select by user input?
		backlightName = files[0].Name()
		log.Printf("Found multiple backlight sources, using: " + backlightName)
	}

	return backlightName, nil
}
