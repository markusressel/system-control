package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseSynclientList parses the output of `synclient -l` into parameter name -> value.
func ParseSynclientList(output string) map[string]string {
	result := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		name, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" || strings.ContainsAny(name, " \t") {
			continue
		}
		result[name] = strings.TrimSpace(value)
	}
	return result
}

// synclientSettings returns the settings of the synaptics driver. An error means
// that the driver is not in use (which is the normal case with libinput).
func synclientSettings() (map[string]string, error) {
	output, err := ExecCommand("synclient", "-l")
	if err != nil {
		return nil, err
	}
	settings := ParseSynclientList(output)
	if len(settings) == 0 {
		return nil, errors.New("synclient returned no settings")
	}
	return settings, nil
}

// GetTouchpadInputDevice returns the name of the touchpad xinput device, if there is one.
func GetTouchpadInputDevice() *string {
	for _, device := range GetInputDevices() {
		if ContainsIgnoreCase(device, "Touchpad") {
			return &device
		}
	}
	return nil
}

// IsTouchpadEnabled reports whether the touchpad is enabled. The synaptics driver and
// libinput (xinput) are both checked, if available. The touchpad counts as enabled
// only if every available backend has it enabled.
func IsTouchpadEnabled() (bool, error) {
	available := false
	enabled := true

	if settings, err := synclientSettings(); err == nil {
		if off, ok := settings["TouchpadOff"]; ok {
			available = true
			enabled = enabled && off == "0"
		}
	}

	if device := GetTouchpadInputDevice(); device != nil {
		available = true
		enabled = enabled && IsInputDeviceEnabled(*device)
	}

	if !available {
		return false, errors.New("no touchpad found")
	}
	return enabled, nil
}

// SetTouchpadEnabled enables or disables the touchpad using every available backend.
// It only fails if no backend could apply the change.
func SetTouchpadEnabled(enabled bool) error {
	off := 1
	if enabled {
		off = 0
	}

	var errs []error
	applied := false

	if _, err := ExecCommand("synclient", "TouchpadOff="+strconv.Itoa(off)); err == nil {
		applied = true
	} else {
		errs = append(errs, err)
	}

	if device := GetTouchpadInputDevice(); device != nil {
		var err error
		if enabled {
			err = EnableInputDevice(*device)
		} else {
			err = DisableInputDevice(*device)
		}
		if err == nil {
			applied = true
		} else {
			errs = append(errs, err)
		}
	} else {
		errs = append(errs, errors.New("no touchpad device found"))
	}

	if applied {
		return nil
	}
	return errors.Join(errs...)
}

// ToggleTouchpadScrollVerticalDirection inverts the vertical scroll direction (natural scrolling).
func ToggleTouchpadScrollVerticalDirection() error {
	settings, err := synclientSettings()
	if err != nil {
		return err
	}
	delta, err := strconv.Atoi(settings["VertScrollDelta"])
	if err != nil {
		return fmt.Errorf("could not read VertScrollDelta: %w", err)
	}

	_, err = ExecCommand("synclient", "VertScrollDelta="+strconv.Itoa(-delta))
	return err
}
