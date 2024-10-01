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

func parseBatteryInfo(file os.DirEntry) (BatteryInfo, error) {
	battery := BatteryInfo{}
	batteryName := file.Name()
	batteryPath := PowerSupplyBasePath + batteryName
	manufacturer, err := ReadTextFromFile(batteryPath + "/manufacturer")
	if err != nil {
		return battery, err
	}
	manufacturer = strings.TrimSpace(manufacturer)

	model, err := ReadTextFromFile(batteryPath + "/model_name")
	if err != nil {
		return battery, err
	}
	model = strings.TrimSpace(model)

	capacityLevel, err := ReadTextFromFile(batteryPath + "/capacity_level")
	if err != nil {
		return battery, err
	}
	capacityLevel = strings.TrimSpace(capacityLevel)

	online, err := ReadTextFromFile(batteryPath + "/online")
	if err != nil {
		return battery, err
	}
	online = strings.TrimSpace(online)

	status, err := ReadTextFromFile(batteryPath + "/status")
	if err != nil {
		return battery, err
	}
	status = strings.TrimSpace(status)

	serialNumber, err := ReadTextFromFile(batteryPath + "/serial_number")
	if err != nil {
		return battery, err
	}
	serialNumber = strings.TrimSpace(serialNumber)

	scope, err := ReadTextFromFile(batteryPath + "/scope")
	if err != nil {
		return battery, err
	}
	scope = strings.TrimSpace(scope)

	bType, err := ReadTextFromFile(batteryPath + "/type")
	if err != nil {
		return battery, err
	}
	bType = strings.TrimSpace(bType)

	battery.Name = batteryName
	battery.Path = batteryPath

	battery.Manufacturer = manufacturer
	battery.Model = model
	battery.CapacityLevel = capacityLevel
	battery.Online = online == "1"
	battery.Status = status
	battery.SerialNumber = serialNumber
	battery.Scope = scope
	battery.Type = bType
	return battery, nil
}

func GetEnergyTarget(battery string) (int64, error) {
	chargeControlEndThreshold := GetChargeControlEndThreshold(battery)
	energyFull, err := GetEnergyFull(battery)
	return int64((float64(energyFull) / 100) * float64(chargeControlEndThreshold)), err
}

func GetChargeControlEndThreshold(battery string) int64 {
	path := PowerSupplyBasePath + battery + "/charge_control_end_threshold"
	value, err := ReadIntFromFile(path)
	if err != nil {
		value = 100
	}
	return value
}

func CalculateRemainingTime(wh int64, w int64) int64 {
	return int64((float64(wh) / float64(w)) * 60 * 60)
}

func IsBatteryCharging(battery string) (bool, error) {
	path := PowerSupplyBasePath + battery + "/status"
	status, err := ReadTextFromFile(path)
	status = strings.TrimSpace(status)
	charging := status == "Charging"
	return charging, err
}

func GetEnergyFull(battery string) (int64, error) {
	path := PowerSupplyBasePath + battery + "/energy_full"
	return ReadIntFromFile(path)
}

func GetEnergyNow(battery string) (int64, error) {
	path := PowerSupplyBasePath + battery + "/energy_now"
	return ReadIntFromFile(path)
}

func GetPowerNow(battery string) (int64, error) {
	path := PowerSupplyBasePath + battery + "/power_now"
	return ReadIntFromFile(path)
}
