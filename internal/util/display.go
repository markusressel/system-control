package util

import (
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

func GetMaxBrightness() int {
	backlightName := findBacklight()
	maxBrightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness
	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(maxBrightness)
}

func GetBrightness() int {
	backlightName := findBacklight()
	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	brightness, err := ReadIntFromFile(brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(brightness)
}

// SetBrightness sets a specific brightness of main the display
func SetBrightness(percentage int) {
	backlightName := findBacklight()
	maxBrightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness
	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness

	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}

	targetValue := int((float32(percentage) / 100.0) * float32(maxBrightness))
	err = WriteIntToFile(targetValue, brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
}

func setBrightnessRaw(backlight string, brightness int) {
	maxBrightness := GetMaxBrightness()
	targetBrightness := brightness
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlight + string(os.PathSeparator) + Brightness

	err := WriteIntToFile(targetBrightness, brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
}

// AdjustBrightness adjusts the brightness of the main display
func AdjustBrightness(change int) {
	backlight := findBacklight()

	maxBrightness := GetMaxBrightness()
	currentBrightness := GetBrightness()

	targetBrightness := currentBrightness + change
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	setBrightnessRaw(backlight, targetBrightness)
}

func findBacklight() string {
	files, err := os.ReadDir(DisplayBacklightPath)
	if err != nil {
		log.Fatal(err)
	}

	var backlightName string
	if len(files) == 0 {
		log.Fatal("No backlight found")
	} else if len(files) == 1 {
		backlightName = files[0].Name()
	} else {
		// TODO: select first? select by user input?
		backlightName = files[0].Name()
		log.Printf("Found multiple backlight sources, using: " + backlightName)
	}

	return backlightName
}
