package util

import (
	"os"
	"strings"
)

type BatteryInfo struct {
	Name string
	Path string

	Manufacturer string
	Model        string
	SerialNumber string
	Scope        string
}

const (
	PowerSupplyBasePath = "/sys/class/power_supply/"
)

func newBatteryInfo(name string) BatteryInfo {
	return BatteryInfo{
		Name: name,
		Path: PowerSupplyBasePath + name,
	}
}

// GetBatteryList returns a list of all batteries found in the system.
func GetBatteryList() (batteryList []BatteryInfo, err error) {
	path := PowerSupplyBasePath
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		battery, err := parseBatteryInfo(file)
		if err != nil {
			continue
		}
		batteryList = append(batteryList, battery)
	}
	return batteryList, nil
}

// parseBatteryInfo parses the battery information from the given directory entry.
// The directory entry should be a directory in the /sys/class/power_supply/ directory.
func parseBatteryInfo(file os.DirEntry) (BatteryInfo, error) {
	batteryName := file.Name()
	battery := newBatteryInfo(batteryName)

	manufacturer, _ := battery.GetManufacturer()
	model, _ := battery.GetModel()
	serialNumber, _ := battery.GetSerialNumber()
	scope, _ := battery.GetScope()

	battery.Manufacturer = manufacturer
	battery.Model = model
	battery.SerialNumber = serialNumber
	battery.Scope = scope

	return battery, nil
}

// GetEnergyTarget returns the target energy level in Wh that the battery should be charged to.
func (battery BatteryInfo) GetEnergyTarget() (int64, error) {
	chargeControlEndThreshold := battery.GetChargeControlEndThreshold()
	energyFull, err := battery.GetEnergyFull()
	return int64((float64(energyFull) / 100) * float64(chargeControlEndThreshold)), err
}

// GetChargeControlEndThreshold returns the charge end threshold in percent.
func (battery BatteryInfo) GetChargeControlEndThreshold() int64 {
	path := battery.Path + "/charge_control_end_threshold"
	value, err := ReadIntFromFile(path)
	if err != nil {
		value = 100
	}
	return value
}

// CalculateRemainingTime calculates the remaining time in seconds until the battery is fully discharged or has reached
// the currently set charge control end threshold.
func CalculateRemainingTime(wh int64, w int64) int64 {
	return int64((float64(wh) / float64(w)) * 60 * 60)
}

func (battery BatteryInfo) GetType() (string, error) {
	path := battery.Path + "/type"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}

func (battery BatteryInfo) IsCharging() (bool, error) {
	path := battery.Path + "/status"
	status, err := ReadTextFromFile(path)
	status = strings.TrimSpace(status)
	charging := EqualsIgnoreCase(status, "Charging")
	return charging, err
}

func (battery BatteryInfo) GetEnergyFull() (int64, error) {
	path := battery.Path + "/energy_full"
	return ReadIntFromFile(path)
}

func (battery BatteryInfo) GetEnergyNow() (int64, error) {
	path := battery.Path + "/energy_now"
	return ReadIntFromFile(path)
}

func (battery BatteryInfo) GetPowerNow() (int64, error) {
	path := battery.Path + "/power_now"
	return ReadIntFromFile(path)
}

func (battery BatteryInfo) GetVoltageNow() (int64, error) {
	path := battery.Path + "/voltage_now"
	rawValue, err := ReadIntFromFile(path)
	if err != nil {
		return 0, err
	}
	return rawValue / 1000000, nil
}

func (battery BatteryInfo) GetVoltageMinDesign() (int64, error) {
	path := battery.Path + "/voltage_min_design"
	rawValue, err := ReadIntFromFile(path)
	if err != nil {
		return 0, err
	}
	return rawValue / 1000000, nil
}

func (battery BatteryInfo) GetTechnology() (string, error) {
	path := battery.Path + "/technology"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}

func (battery BatteryInfo) GetCycleCount() (int64, error) {
	path := battery.Path + "/cycle_count"
	return ReadIntFromFile(path)
}

func (battery BatteryInfo) GetCapacity() (int64, error) {
	path := battery.Path + "/capacity"
	return ReadIntFromFile(path)
}

func (battery BatteryInfo) GetCapacityLevel() (string, error) {
	path := battery.Path + "/capacity_level"
	capacityLevel, err := ReadTextFromFile(path)
	if err != nil {
		return capacityLevel, err
	}
	capacityLevel = strings.TrimSpace(capacityLevel)
	return capacityLevel, nil
}

func (battery BatteryInfo) GetStatus() (string, error) {
	path := battery.Path + "/status"
	status, err := ReadTextFromFile(path)
	if err != nil {
		return status, err
	}
	status = strings.TrimSpace(status)
	return status, nil
}

func (battery BatteryInfo) IsOnline() (bool, error) {
	path := battery.Path + "/online"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(rawValue) == "1", nil
}

func (battery BatteryInfo) IsPresent() (bool, error) {
	path := battery.Path + "/present"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(rawValue) == "1", nil
}

func (battery BatteryInfo) GetScope() (string, error) {
	path := battery.Path + "/scope"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}

func (battery BatteryInfo) GetSerialNumber() (string, error) {
	path := battery.Path + "/serial_number"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}

func (battery BatteryInfo) GetManufacturer() (string, error) {
	path := battery.Path + "/manufacturer"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}

func (battery BatteryInfo) GetModel() (string, error) {
	path := battery.Path + "/model_name"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}
