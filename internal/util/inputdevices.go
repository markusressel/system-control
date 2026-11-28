package util

import (
	"errors"
	"math"
	"os"
	"regexp"
	"strings"
)

func findKeyboardBacklight() (string, error) {
	files, err := os.ReadDir(LedsPath)
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(".*(kbd|keyboard).*")
	for _, f := range files {
		if r.MatchString(f.Name()) {
			return f.Name(), nil
		}
	}

	return "", errors.New("no keyboard backlight found")
}

func GetKeyboardBrightness() (int, error) {
	backlightName, err := findKeyboardBacklight()
	if err != nil {
		return -1, err
	}
	brightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	brightness, err := ReadIntFromFile(brightnessPath)
	if err != nil {
		return -1, err
	}
	return int(brightness), nil
}

func SetKeyboardBrightness(brightness int) (int, error) {
	backlightName, err := findKeyboardBacklight()
	if err != nil {
		return -1, err
	}
	brightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	maxBrightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness

	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		return -1, err
	}

	targetValue := math.Max(0, math.Min(float64(maxBrightness), float64(brightness)))
	err = WriteIntToFile(int(targetValue), brightnessPath)
	if err != nil {
		return -1, err
	}
	return int(targetValue), nil
}

func GetInputDevices() []string {
	result, _ := ExecCommand("xinput", "list", "--name-only")
	inputDeviceNames := strings.Split(result, "\n")
	// strip leading "~ " from each entry, which indicates "disabled" state
	for idx, inputDeviceName := range inputDeviceNames {
		inputDeviceName = strings.TrimLeft(inputDeviceName, "∼ ")
		inputDeviceName = strings.TrimSpace(inputDeviceName)
		inputDeviceNames[idx] = inputDeviceName
	}
	return inputDeviceNames
}

func IsInputDeviceEnabled(name string) bool {
	result, _ := ExecCommand("xinput", "list", name)
	return !ContainsIgnoreCase(result, "This device is disabled")
}

func EnableInputDevice(name string) error {
	_, err := ExecCommand("xinput", "enable", name)
	return err
}

func DisableInputDevice(name string) error {
	_, err := ExecCommand("xinput", "disable", name)
	return err
}
