package util

import (
	"errors"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func findKeyboardBacklight() string {
	files, err := os.ReadDir(LedsPath)
	if err != nil {
		log.Fatal(err)
	}

	var kbdBacklight string
	r := regexp.MustCompile(".*(kbd|keyboard).*")
	for _, f := range files {
		if r.MatchString(f.Name()) {
			return f.Name()
		}
	}

	log.Fatal("No keyboard backlight found")
	return kbdBacklight
}

func GetKeyboardBrightness() int {
	backlightName := findKeyboardBacklight()
	brightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	brightness, err := ReadIntFromFile(brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(brightness)
}

func SetKeyboardBrightness(brightness int) int {
	backlightName := findKeyboardBacklight()
	brightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	maxBrightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness

	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}

	targetValue := math.Max(0, math.Min(float64(maxBrightness), float64(brightness)))
	err = WriteIntToFile(int(targetValue), brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(targetValue)
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

func GetTouchpadInputDevice() *string {
	inputDevices := GetInputDevices()
	for _, device := range inputDevices {
		if ContainsIgnoreCase(device, "Touchpad") {
			return &device
		}
	}

	return nil
}

func IsTouchpadEnabledLibinput() bool {
	touchpadDevice := GetTouchpadInputDevice()
	if touchpadDevice != nil {
		return IsInputDeviceEnabled(*touchpadDevice)
	} else {
		return false
	}
}

func IsTouchpadEnabledSynaptics() bool {
	result, _ := ExecCommand("synclient")
	regex := regexp.MustCompile("\\s*TouchpadOff\\s*=\\s*(\\d)")

	submatch := regex.FindStringSubmatch(result)[0]
	submatch = strings.TrimSpace(submatch)
	value := submatch[len(submatch)-1:]

	resultInt, _ := strconv.Atoi(value)
	return resultInt == 0
}

func IsTouchpadEnabled() bool {
	return IsTouchpadEnabledSynaptics() && IsTouchpadEnabledLibinput()
}

func SetTouchpadEnabled(enabled bool) error {
	err := SetTouchpadEnabledSynaptics(enabled)
	if err != nil {
		return err
	}
	return SetTouchpadEnabledLibinput(enabled)
}

func SetTouchpadEnabledSynaptics(enabled bool) error {
	var enabledInt int
	if enabled {
		enabledInt = 0
	} else {
		enabledInt = 1
	}

	_, err := ExecCommand("synclient", "TouchpadOff="+strconv.Itoa(enabledInt))
	return err
}

func SetTouchpadEnabledLibinput(enabled bool) (err error) {
	touchpadDevice := GetTouchpadInputDevice()
	if touchpadDevice != nil {
		if enabled {
			err = EnableInputDevice(*touchpadDevice)
		} else {
			err = DisableInputDevice(*touchpadDevice)
		}
	} else {
		err = errors.New("no touchpad device found")
	}

	return err
}
