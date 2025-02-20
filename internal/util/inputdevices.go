package util

import (
	"errors"
	"math"
	"os"
	"regexp"
	"strconv"
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

func ToggleTouchpadScrollVerticalDirection() error {
	synclientProperties, err := GetSynclientProperties()
	if err != nil {
		return err
	}

	_, err = ExecCommand("synclient",
		"-l",
		"VertScrollDelta="+strconv.Itoa(synclientProperties.VertScrollDelta*-1),
	)
	return err
}

type SynclientProperties struct {
	LeftEdge                int
	RightEdge               int
	TopEdge                 int
	BottomEdge              int
	FingerLow               int
	FingerHigh              int
	MaxTapTime              int
	MaxTapMove              int
	MaxDoubleTapTime        int
	SingleTapTimeout        int
	ClickTime               int
	EmulateMidButtonTime    int
	EmulateTwoFingerMinZ    int
	EmulateTwoFingerMinW    int
	VertScrollDelta         int
	HorizScrollDelta        int
	VertEdgeScroll          int
	HorizEdgeScroll         int
	CornerCoasting          int
	VertTwoFingerScroll     int
	HorizTwoFingerScroll    int
	MinSpeed                int
	MaxSpeed                float64
	AccelFactor             float64
	TouchpadOff             int
	LockedDrags             int
	LockedDragTimeout       int
	RTCornerButton          int
	RBCornerButton          int
	LTCornerButton          int
	LBCornerButton          int
	TapButton1              int
	TapButton2              int
	TapButton3              int
	ClickFinger1            int
	ClickFinger2            int
	ClickFinger3            int
	CircularScrolling       int
	CircScrollDelta         float64
	CircScrollTrigger       int
	CircularPad             int
	PalmDetect              int
	PalmMinWidth            int
	PalmMinZ                int
	CoastingSpeed           int
	CoastingFriction        int
	PressureMotionMinZ      int
	PressureMotionMaxZ      int
	PressureMotionMinFactor int
	PressureMotionMaxFactor int
	GrabEventDevice         int
	TapAndDragGesture       int
	AreaLeftEdge            int
	AreaRightEdge           int
	AreaTopEdge             int
	AreaBottomEdge          int
	HorizHysteresis         int
	VertHysteresis          int
	ClickPad                int
	RightButtonAreaLeft     int
	RightButtonAreaRight    int
	RightButtonAreaTop      int
	RightButtonAreaBottom   int
	MiddleButtonAreaLeft    int
	MiddleButtonAreaRight   int
	MiddleButtonAreaTop     int
	MiddleButtonAreaBottom  int
}

// Parameter settings:
//
//	LeftEdge                = 159
//	RightEdge               = 3831
//	TopEdge                 = 139
//	BottomEdge              = 2439
//	FingerLow               = 25
//	FingerHigh              = 30
//	MaxTapTime              = 180
//	MaxTapMove              = 209
//	MaxDoubleTapTime        = 180
//	SingleTapTimeout        = 180
//	ClickTime               = 100
//	EmulateMidButtonTime    = 0
//	EmulateTwoFingerMinZ    = 282
//	EmulateTwoFingerMinW    = 7
//	VertScrollDelta         = 95
//	HorizScrollDelta        = 95
//	VertEdgeScroll          = 0
//	HorizEdgeScroll         = 0
//	CornerCoasting          = 0
//	VertTwoFingerScroll     = 1
//	HorizTwoFingerScroll    = 1
//	MinSpeed                = 1
//	MaxSpeed                = 1.75
//	AccelFactor             = 0.03
//	TouchpadOff             = 1
//	LockedDrags             = 0
//	LockedDragTimeout       = 5000
//	RTCornerButton          = 0
//	RBCornerButton          = 0
//	LTCornerButton          = 0
//	LBCornerButton          = 0
//	TapButton1              = 0
//	TapButton2              = 0
//	TapButton3              = 0
//	ClickFinger1            = 1
//	ClickFinger2            = 3
//	ClickFinger3            = 2
//	CircularScrolling       = 0
//	CircScrollDelta         = 0.1
//	CircScrollTrigger       = 0
//	CircularPad             = 0
//	PalmDetect              = 0
//	PalmMinWidth            = 10
//	PalmMinZ                = 200
//	CoastingSpeed           = 20
//	CoastingFriction        = 50
//	PressureMotionMinZ      = 30
//	PressureMotionMaxZ      = 160
//	PressureMotionMinFactor = 1
//	PressureMotionMaxFactor = 1
//	GrabEventDevice         = 0
//	TapAndDragGesture       = 1
//	AreaLeftEdge            = 0
//	AreaRightEdge           = 0
//	AreaTopEdge             = 0
//	AreaBottomEdge          = 0
//	HorizHysteresis         = 23
//	VertHysteresis          = 23
//	ClickPad                = 1
//	RightButtonAreaLeft     = 1995
//	RightButtonAreaRight    = 0
//	RightButtonAreaTop      = 2113
//	RightButtonAreaBottom   = 0
//	MiddleButtonAreaLeft    = 0
//	MiddleButtonAreaRight   = 0
//	MiddleButtonAreaTop     = 0
//	MiddleButtonAreaBottom  = 0
func GetSynclientProperties() (*SynclientProperties, error) {
	result, err := ExecCommand("synclient", "-l")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(result, "\n")
	properties := &SynclientProperties{}
	for _, line := range lines {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			switch key {
			case "LeftEdge":
				properties.LeftEdge, _ = strconv.Atoi(value)
			case "RightEdge":
				properties.RightEdge, _ = strconv.Atoi(value)
			case "TopEdge":
				properties.TopEdge, _ = strconv.Atoi(value)
			case "BottomEdge":
				properties.BottomEdge, _ = strconv.Atoi(value)
			case "FingerLow":
				properties.FingerLow, _ = strconv.Atoi(value)
			case "FingerHigh":
				properties.FingerHigh, _ = strconv.Atoi(value)
			case "MaxTapTime":
				properties.MaxTapTime, _ = strconv.Atoi(value)
			case "MaxTapMove":
				properties.MaxTapMove, _ = strconv.Atoi(value)
			case "MaxDoubleTapTime":
				properties.MaxDoubleTapTime, _ = strconv.Atoi(value)
			case "SingleTapTimeout":
				properties.SingleTapTimeout, _ = strconv.Atoi(value)
			case "ClickTime":
				properties.ClickTime, _ = strconv.Atoi(value)
			case "EmulateMidButtonTime":
				properties.EmulateMidButtonTime, _ = strconv.Atoi(value)
			case "EmulateTwoFingerMinZ":
				properties.EmulateTwoFingerMinZ, _ = strconv.Atoi(value)
			case "EmulateTwoFingerMinW":
				properties.EmulateTwoFingerMinW, _ = strconv.Atoi(value)
			case "VertScrollDelta":
				properties.VertScrollDelta, _ = strconv.Atoi(value)
			case "HorizScrollDelta":
				properties.HorizScrollDelta, _ = strconv.Atoi(value)
			case "VertEdgeScroll":
				properties.VertEdgeScroll, _ = strconv.Atoi(value)
			case "HorizEdgeScroll":
				properties.HorizEdgeScroll, _ = strconv.Atoi(value)
			case "CornerCoasting":
				properties.CornerCoasting, _ = strconv.Atoi(value)
			case "VertTwoFingerScroll":
				properties.VertTwoFingerScroll, _ = strconv.Atoi(value)
			case "HorizTwoFingerScroll":
				properties.HorizTwoFingerScroll, _ = strconv.Atoi(value)
			case "MinSpeed":
				properties.MinSpeed, _ = strconv.Atoi(value)
			case "MaxSpeed":
				properties.MaxSpeed, _ = strconv.ParseFloat(value, 64)
			case "AccelFactor":
				properties.AccelFactor, _ = strconv.ParseFloat(value, 64)
			case "TouchpadOff":
				properties.TouchpadOff, _ = strconv.Atoi(value)
			case "LockedDrags":
				properties.LockedDrags, _ = strconv.Atoi(value)
			case "LockedDragTimeout":
				properties.LockedDragTimeout, _ = strconv.Atoi(value)
			case "RTCornerButton":
				properties.RTCornerButton, _ = strconv.Atoi(value)
			case "RBCornerButton":
				properties.RBCornerButton, _ = strconv.Atoi(value)
			case "LTCornerButton":
				properties.LTCornerButton, _ = strconv.Atoi(value)
			case "LBCornerButton":
				properties.LBCornerButton, _ = strconv.Atoi(value)
			case "TapButton1":
				properties.TapButton1, _ = strconv.Atoi(value)
			case "TapButton2":
				properties.TapButton2, _ = strconv.Atoi(value)
			case "TapButton3":
				properties.TapButton3, _ = strconv.Atoi(value)
			case "ClickFinger1":
				properties.ClickFinger1, _ = strconv.Atoi(value)
			case "ClickFinger2":
				properties.ClickFinger2, _ = strconv.Atoi(value)
			case "ClickFinger3":
				properties.ClickFinger3, _ = strconv.Atoi(value)
			case "CircularScrolling":
				properties.CircularScrolling, _ = strconv.Atoi(value)
			case "CircScrollDelta":
				properties.CircScrollDelta, _ = strconv.ParseFloat(value, 64)
			case "CircScrollTrigger":
				properties.CircScrollTrigger, _ = strconv.Atoi(value)
			case "CircularPad":
				properties.CircularPad, _ = strconv.Atoi(value)
			case "PalmDetect":
				properties.PalmDetect, _ = strconv.Atoi(value)
			case "PalmMinWidth":
				properties.PalmMinWidth, _ = strconv.Atoi(value)
			case "PalmMinZ":
				properties.PalmMinZ, _ = strconv.Atoi(value)
			case "CoastingSpeed":
				properties.CoastingSpeed, _ = strconv.Atoi(value)
			case "CoastingFriction":
				properties.CoastingFriction, _ = strconv.Atoi(value)
			case "PressureMotionMinZ":
				properties.PressureMotionMinZ, _ = strconv.Atoi(value)
			case "PressureMotionMaxZ":
				properties.PressureMotionMaxZ, _ = strconv.Atoi(value)
			case "PressureMotionMinFactor":
				properties.PressureMotionMinFactor, _ = strconv.Atoi(value)
			case "PressureMotionMaxFactor":
				properties.PressureMotionMaxFactor, _ = strconv.Atoi(value)
			case "GrabEventDevice":
				properties.GrabEventDevice, _ = strconv.Atoi(value)
			case "TapAndDragGesture":
				properties.TapAndDragGesture, _ = strconv.Atoi(value)
			case "AreaLeftEdge":
				properties.AreaLeftEdge, _ = strconv.Atoi(value)
			case "AreaRightEdge":
				properties.AreaRightEdge, _ = strconv.Atoi(value)
			case "AreaTopEdge":
				properties.AreaTopEdge, _ = strconv.Atoi(value)
			case "AreaBottomEdge":
				properties.AreaBottomEdge, _ = strconv.Atoi(value)
			case "HorizHysteresis":
				properties.HorizHysteresis, _ = strconv.Atoi(value)
			case "VertHysteresis":
				properties.VertHysteresis, _ = strconv.Atoi(value)
			case "ClickPad":
				properties.ClickPad, _ = strconv.Atoi(value)
			case "RightButtonAreaLeft":
				properties.RightButtonAreaLeft, _ = strconv.Atoi(value)
			case "RightButtonAreaRight":
				properties.RightButtonAreaRight, _ = strconv.Atoi(value)
			case "RightButtonAreaTop":
				properties.RightButtonAreaTop, _ = strconv.Atoi(value)
			case "RightButtonAreaBottom":
				properties.RightButtonAreaBottom, _ = strconv.Atoi(value)
			case "MiddleButtonAreaLeft":
				properties.MiddleButtonAreaLeft, _ = strconv.Atoi(value)
			case "MiddleButtonAreaRight":
				properties.MiddleButtonAreaRight, _ = strconv.Atoi(value)
			case "MiddleButtonAreaTop":
				properties.MiddleButtonAreaTop, _ = strconv.Atoi(value)
			case "MiddleButtonAreaBottom":
				properties.MiddleButtonAreaBottom, _ = strconv.Atoi(value)
			}
		}
	}

	return properties, nil
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
