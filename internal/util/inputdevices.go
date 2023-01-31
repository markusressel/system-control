package util

import (
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
	return strings.Split(result, "\n")
}

func IsInputDeviceEnabled(name string) bool {
	result, _ := ExecCommand("xinput", "list", name)
	return !ContainsIgnoreCase(result, "This device is disabled")
}

func EnableInputDevice(name string) {
	_, err := ExecCommand("xinput", "enable", name)
	if err != nil {
		log.Fatal(err)
	}
}

func DisableInputDevice(name string) {
	_, err := ExecCommand("xinput", "disable", name)
	if err != nil {
		log.Fatal(err)
	}
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

func SetTouchpadEnabled(enabled bool) {
	SetTouchpadEnabledSynaptics(enabled)
	SetTouchpadEnabledLibinput(enabled)
}

func SetTouchpadEnabledSynaptics(enabled bool) {
	var enabledInt int
	if enabled {
		enabledInt = 0
	} else {
		enabledInt = 1
	}

	_, err := ExecCommand("synclient", "TouchpadOff="+strconv.Itoa(enabledInt))
	if err != nil {
		log.Fatal(err)
	}
}

func SetTouchpadEnabledLibinput(enabled bool) {
	touchpadDevice := GetTouchpadInputDevice()
	if touchpadDevice != nil {
		if enabled {
			EnableInputDevice(*touchpadDevice)
		} else {
			DisableInputDevice(*touchpadDevice)
		}
	} else {
		log.Fatal("no touchpad device found")
	}
}
