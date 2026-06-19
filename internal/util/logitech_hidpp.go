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
	res, err := transactHIDPP(f, 0x00, 0x00, []byte{0x10, 0x00})
	if err != nil {
		return 0, "", "", fmt.Errorf("dynamic feature lookup failed: %w", err)
	}

	featureIdx := res[4]
	if featureIdx == 0 {
		return 0, "", "", errors.New("feature 0x1000 (battery unified level status) not supported by this device")
	}

	// Step 2: Query Battery Level Status (Function 0x00 of Battery Feature)
	resBatt, err := transactHIDPP(f, featureIdx, 0x00, nil)
	if err != nil {
		return 0, "", "", fmt.Errorf("battery query failed: %w", err)
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

// transactHIDPP builds a 20-byte long report (0x11) for device 0xFF, writes it to the hidraw file,
// and reads the response, ensuring the response matches the expected feature index and function ID.
func transactHIDPP(f *os.File, featureIdx, funcID byte, data []byte) ([]byte, error) {
	req := make([]byte, 20)
	req[0] = 0x11
	req[1] = 0xFF
	req[2] = featureIdx
	req[3] = funcID
	if len(data) > 0 {
		copy(req[4:], data)
	}
	return writeAndReadHIDPP(f, req, featureIdx, funcID)
}

// writeAndReadHIDPP writes a request to the hidraw device and reads the response,
// filtering for the matching Report ID (0x11), Device Index, Feature Index, and Function ID,
// and handling HID++ error packets.
func writeAndReadHIDPP(f *os.File, req []byte, expectedFeatureIdx, expectedFuncID byte) ([]byte, error) {
	if _, err := f.Write(req); err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	deadline := time.Now().Add(1500 * time.Millisecond)
	for {
		if time.Now().After(deadline) {
			return nil, errors.New("timeout waiting for response")
		}

		err := f.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		if err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}

		buf := make([]byte, 64)
		n, err := f.Read(buf)
		if err != nil {
			if os.IsTimeout(err) {
				continue
			}
			return nil, fmt.Errorf("read error: %w", err)
		}

		if n > 0 {
			// Expect long report (0x11)
			if n >= 5 && buf[0] == 0x11 {
				// Check for error packet: buf[2] == 0xFF and matching feature index at buf[3]
				if buf[2] == 0xFF && buf[3] == expectedFeatureIdx {
					return nil, fmt.Errorf("device returned protocol error: code %d", buf[5])
				}
				// Verify feature index and function ID match
				if buf[2] == expectedFeatureIdx && buf[3] == expectedFuncID {
					return buf[:n], nil
				}
			}
		}
	}
}
