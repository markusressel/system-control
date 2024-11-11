package util

import (
	"os"
	"strings"
)

type BatteryInfo struct {
	Name string
	Path string

	Manufacturer  string
	Model         string
	Capacity      int64
	CapacityLevel string
	Online        bool
	// f.ex. "Charging" or "Discharging"
	Status       string
	SerialNumber string
	Scope        string
	Type         string
}

const (
	PowerSupplyBasePath = "/sys/class/power_supply/"
)

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
	battery := BatteryInfo{}
	batteryName := file.Name()
	batteryPath := PowerSupplyBasePath + batteryName
	manufacturer, _ := ReadTextFromFile(batteryPath + "/manufacturer")
	manufacturer = strings.TrimSpace(manufacturer)

	model, _ := ReadTextFromFile(batteryPath + "/model_name")
	model = strings.TrimSpace(model)

	capacity, _ := ReadIntFromFile(batteryPath + "/capacity")

	capacityLevel, _ := ReadTextFromFile(batteryPath + "/capacity_level")
	capacityLevel = strings.TrimSpace(capacityLevel)

	online, _ := ReadTextFromFile(batteryPath + "/online")
	online = strings.TrimSpace(online)

	status, _ := ReadTextFromFile(batteryPath + "/status")
	status = strings.TrimSpace(status)

	serialNumber, _ := ReadTextFromFile(batteryPath + "/serial_number")
	serialNumber = strings.TrimSpace(serialNumber)

	scope, _ := ReadTextFromFile(batteryPath + "/scope")
	scope = strings.TrimSpace(scope)

	bType, _ := ReadTextFromFile(batteryPath + "/type")
	bType = strings.TrimSpace(bType)

	battery.Name = batteryName
	battery.Path = batteryPath

	battery.Manufacturer = manufacturer
	battery.Model = model
	battery.Capacity = capacity
	battery.CapacityLevel = capacityLevel
	battery.Online = online == "1"
	battery.Status = status
	battery.SerialNumber = serialNumber
	battery.Scope = scope
	battery.Type = bType
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

func (battery BatteryInfo) IsCharging() (bool, error) {
	path := battery.Path + "/status"
	status, err := ReadTextFromFile(path)
	status = strings.TrimSpace(status)
	charging := status == "Charging"
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
