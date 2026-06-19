package util

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type BatteryInfo struct {
	Name string
	Path string

	Manufacturer string
	Model        string
	SerialNumber string
	Scope        string

	// Cached HID++ values
	hidppQueried       bool
	hidppCapacity      int64
	hidppCapacityLevel string
	hidppStatus        string
	hidppErr           error
	hidppIsCached      bool
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
func (battery BatteryInfo) GetEnergyTarget() (float64, error) {
	chargeControlEndThreshold := battery.GetChargeControlEndThreshold()
	energyFull, err := battery.GetEnergyFull()
	return (float64(energyFull) / 100) * float64(chargeControlEndThreshold), err
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
func CalculateRemainingTime(wh float64, w float64) int64 {
	return int64((wh / w) * 60 * 60)
}

func (battery BatteryInfo) GetType() (string, error) {
	path := battery.Path + "/type"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}

func (battery *BatteryInfo) IsCharging() (bool, error) {
	status, err := battery.GetStatus()
	charging := EqualsIgnoreCase(status, "Charging")
	return charging, err
}

// GetEnergyFull returns the energy level of the battery in Wh when fully charged.
func (battery BatteryInfo) GetEnergyFull() (float64, error) {
	path := battery.Path + "/energy_full"
	rawValue, err := ReadIntFromFile(path)
	scaledValue := float64(rawValue) / 1000000
	if err != nil {
		return scaledValue, err
	}
	return scaledValue, nil
}

// GetEnergyFullDesign returns the design energy level of the battery in Wh.
func (battery BatteryInfo) GetEnergyFullDesign() (float64, error) {
	path := battery.Path + "/energy_full_design"
	rawValue, err := ReadIntFromFile(path)
	scaledValue := float64(rawValue) / 1000000
	if err != nil {
		return scaledValue, err
	}
	return scaledValue, nil
}

// GetEnergyNow returns the current energy level of the battery in Wh.
func (battery BatteryInfo) GetEnergyNow() (float64, error) {
	path := battery.Path + "/energy_now"
	rawValue, err := ReadIntFromFile(path)
	scaledValue := float64(rawValue) / 1000000
	if err != nil {
		return scaledValue, err
	}
	return scaledValue, nil
}

// GetPowerNow returns the current power usage of the battery in Watts.
func (battery BatteryInfo) GetPowerNow() (float64, error) {
	path := battery.Path + "/power_now"
	rawValue, err := ReadIntFromFile(path)
	scaledValue := float64(rawValue) / 1000000
	if err != nil {
		return scaledValue, err
	}
	return scaledValue, nil
}

// GetVoltageNow returns the current voltage of the battery in Volts.
func (battery BatteryInfo) GetVoltageNow() (float64, error) {
	path := battery.Path + "/voltage_now"
	rawValue, err := ReadIntFromFile(path)
	scaledValue := float64(rawValue) / 1000000
	if err != nil {
		return scaledValue, err
	}
	return scaledValue, nil
}

// GetVoltageMinDesign returns the minimum voltage of the battery in Volts.
func (battery BatteryInfo) GetVoltageMinDesign() (float64, error) {
	path := battery.Path + "/voltage_min_design"
	rawValue, err := ReadIntFromFile(path)
	scaledValue := float64(rawValue) / 1000000
	if err != nil {
		return scaledValue, err
	}
	return scaledValue, nil
}

// GetTechnology returns the technology of the battery. For example, "Li-ion", "Li-poly", etc.
func (battery BatteryInfo) GetTechnology() (string, error) {
	path := battery.Path + "/technology"
	rawValue, err := ReadTextFromFile(path)
	if err != nil {
		return rawValue, err
	}
	return strings.TrimSpace(rawValue), nil
}

// GetCycleCount returns the current cycle count of the battery.
func (battery BatteryInfo) GetCycleCount() (int64, error) {
	path := battery.Path + "/cycle_count"
	return ReadIntFromFile(path)
}

type BatteryCache struct {
	Capacity      int64  `json:"capacity"`
	CapacityLevel string `json:"capacity_level"`
	Status        string `json:"status"`
}

func getPersistenceDir() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(usr.HomeDir, ".config", "system-control", "persistence")
}

func (battery *BatteryInfo) getCacheFile() string {
	dir := getPersistenceDir()
	if dir == "" {
		return ""
	}
	key := battery.SerialNumber
	if key == "" {
		key = battery.Name
	}
	return filepath.Join(dir, "battery_cache_"+key+".sav")
}

func (battery *BatteryInfo) saveToCache() {
	file := battery.getCacheFile()
	if file == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(file), 0755)
	cache := BatteryCache{
		Capacity:      battery.hidppCapacity,
		CapacityLevel: battery.hidppCapacityLevel,
		Status:        battery.hidppStatus,
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err == nil {
		_ = os.WriteFile(file, data, 0644)
	}
}

func (battery *BatteryInfo) loadFromCache() {
	file := battery.getCacheFile()
	if file == "" {
		return
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return
	}
	var cache BatteryCache
	if err := json.Unmarshal(data, &cache); err == nil {
		battery.hidppCapacity = cache.Capacity
		battery.hidppCapacityLevel = cache.CapacityLevel
		battery.hidppStatus = cache.Status
		battery.hidppErr = nil
		battery.hidppIsCached = true
	}
}

// IsCached returns true if the battery info was loaded from the persistence cache.
func (battery *BatteryInfo) IsCached() bool {
	battery.ResolveHIDPP()
	return battery.hidppIsCached
}

// ResolveHIDPP queries the Logitech device if applicable and caches the result.
func (battery *BatteryInfo) ResolveHIDPP() {
	if battery.hidppQueried {
		return
	}
	battery.hidppQueried = true
	if strings.ToLower(battery.Manufacturer) != "logitech" {
		return
	}

	hidrawPath, err := battery.GetHidrawPath()
	if err != nil {
		battery.hidppErr = err
		battery.loadFromCache()
		return
	}

	level, capLevel, status, err := QueryLogitechBatteryHIDPP(hidrawPath)
	if err != nil {
		battery.hidppErr = err
		battery.loadFromCache()
		return
	}

	battery.hidppCapacity = level
	battery.hidppCapacityLevel = capLevel
	battery.hidppStatus = status
	battery.hidppErr = nil
	battery.hidppIsCached = false
	battery.saveToCache()
}

func (battery *BatteryInfo) useHIDPP() bool {
	battery.ResolveHIDPP()
	return battery.hidppErr == nil && battery.hidppQueried && strings.ToLower(battery.Manufacturer) == "logitech"
}

// GetCapacity returns the current battery capacity in percent.
func (battery *BatteryInfo) GetCapacity() (int64, error) {
	if battery.useHIDPP() {
		return battery.hidppCapacity, nil
	}
	path := battery.Path + "/capacity"
	return ReadIntFromFile(path)
}

// GetCapacityLevel returns the current capacity level of the battery.
func (battery *BatteryInfo) GetCapacityLevel() (string, error) {
	if battery.useHIDPP() {
		return battery.hidppCapacityLevel, nil
	}
	path := battery.Path + "/capacity_level"
	capacityLevel, err := ReadTextFromFile(path)
	if err != nil {
		return capacityLevel, err
	}
	capacityLevel = strings.TrimSpace(capacityLevel)
	return capacityLevel, nil
}

// GetStatus returns the current status of the battery. For example, "Charging", "Discharging", "Not Charging", etc.
func (battery *BatteryInfo) GetStatus() (string, error) {
	if battery.useHIDPP() {
		return battery.hidppStatus, nil
	}
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

// GetDegradation returns the current battery degradation in percent.
func (battery BatteryInfo) GetDegradation() (float64, error) {
	energyFull, err := battery.GetEnergyFull()
	energyFullDesign, err := battery.GetEnergyFullDesign()
	if err != nil {
		return 0, err
	}
	return (1 - (float64(energyFull) / float64(energyFullDesign))) * 100, nil
}
