package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetHidrawPath finds the raw HID device path corresponding to this battery, if any.
func (battery *BatteryInfo) GetHidrawPath() (string, error) {
	hidrawDir := filepath.Join(battery.Path, "device", "hidraw")
	files, err := os.ReadDir(hidrawDir)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "hidraw") {
			return filepath.Join("/dev", file.Name()), nil
		}
	}
	return "", fmt.Errorf("no hidraw device found for battery %s", battery.Name)
}

// QueryLogitechBatteryHIDPP queries the battery level and status directly from the raw hidraw device using HID++ 2.0.
func QueryLogitechBatteryHIDPP(hidrawPath string) (level int64, levelText string, status string, err error) {
	f, err := os.OpenFile(hidrawPath, os.O_RDWR, 0)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to open device: %w", err)
	}
	defer f.Close()

	// Step 1: Query Feature ID 0x1000 index using Root Feature (0x0000)
	// Long report (0x11), Device direct/broadcast (0xFF), Root feature index (0x00), GetFeature function (0x00)
	req := make([]byte, 20)
	req[0] = 0x11
	req[1] = 0xFF
	req[2] = 0x00
	req[3] = 0x00
	req[4] = 0x10 // Feature ID 0x1000 MSB
	req[5] = 0x00 // LSB

	if _, err := f.Write(req); err != nil {
		return 0, "", "", fmt.Errorf("failed to write GetFeature query: %w", err)
	}

	// Read response with timeout
	var res []byte
	deadline := time.Now().Add(1500 * time.Millisecond)
	for {
		if time.Now().After(deadline) {
			return 0, "", "", errors.New("timeout waiting for GetFeature response")
		}

		err = f.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		if err != nil {
			return 0, "", "", fmt.Errorf("failed to set read deadline: %w", err)
		}

		buf := make([]byte, 64)
		n, err := f.Read(buf)
		if err != nil {
			if os.IsTimeout(err) {
				continue
			}
			return 0, "", "", fmt.Errorf("failed to read from device: %w", err)
		}

		if n > 0 {
			// Expect long report (0x11)
			if n >= 5 && buf[0] == 0x11 {
				// Check for error packet: buf[2] == 0xFF
				if buf[2] == 0xFF && buf[3] == 0x00 {
					return 0, "", "", fmt.Errorf("device returned error for dynamic feature lookup: code %d", buf[5])
				}
				// Verify feature index 0x00, function 0x00
				if buf[2] == 0x00 && buf[3] == 0x00 {
					res = buf[:n]
					break
				}
			}
		}
	}

	featureIdx := res[4]
	if featureIdx == 0 {
		return 0, "", "", errors.New("feature 0x1000 (battery unified level status) not supported by this device")
	}

	// Step 2: Query Battery Level Status (Function 0x00 of Battery Feature)
	reqBatt := make([]byte, 20)
	reqBatt[0] = 0x11
	reqBatt[1] = 0xFF
	reqBatt[2] = featureIdx
	reqBatt[3] = 0x00 // GetBatteryLevelStatus

	if _, err := f.Write(reqBatt); err != nil {
		return 0, "", "", fmt.Errorf("failed to write battery query: %w", err)
	}

	// Read response with timeout
	var resBatt []byte
	deadline = time.Now().Add(1500 * time.Millisecond)
	for {
		if time.Now().After(deadline) {
			return 0, "", "", errors.New("timeout waiting for battery level response")
		}

		err = f.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		if err != nil {
			return 0, "", "", fmt.Errorf("failed to set read deadline: %w", err)
		}

		buf := make([]byte, 64)
		n, err := f.Read(buf)
		if err != nil {
			if os.IsTimeout(err) {
				continue
			}
			return 0, "", "", fmt.Errorf("failed to read from device: %w", err)
		}

		if n > 0 {
			// Expect long report (0x11) from battery feature index
			if n >= 7 && buf[0] == 0x11 {
				if buf[2] == 0xFF && buf[3] == featureIdx {
					return 0, "", "", fmt.Errorf("device returned error for battery query: code %d", buf[5])
				}
				if buf[2] == featureIdx && buf[3] == 0x00 {
					resBatt = buf[:n]
					break
				}
			}
		}
	}

	levelVal := int64(resBatt[4])
	statusVal := resBatt[6]

	statusMap := map[byte]string{
		0: "Discharging",
		1: "Charging",
		2: "In Between",
		3: "Charged",
		4: "Low",
		5: "Invalid/Error",
	}
	statusStr := statusMap[statusVal]
	if statusStr == "" {
		statusStr = fmt.Sprintf("Unknown (%d)", statusVal)
	}

	var capacityLevel string
	switch {
	case levelVal >= 80:
		capacityLevel = "Full"
	case levelVal >= 25:
		capacityLevel = "Normal"
	case levelVal >= 10:
		capacityLevel = "Low"
	default:
		capacityLevel = "Critical"
	}

	return levelVal, capacityLevel, statusStr, nil
}
