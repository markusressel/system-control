package battery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakePowerSupply creates a sysfs like directory with the given attributes and returns its base path.
func fakePowerSupply(t *testing.T, supplies map[string]map[string]string) {
	base := t.TempDir() + "/"
	for name, attributes := range supplies {
		dir := filepath.Join(base, name)
		assert.NoError(t, os.MkdirAll(dir, 0755))
		for attribute, value := range attributes {
			assert.NoError(t, os.WriteFile(filepath.Join(dir, attribute), []byte(value+"\n"), 0644))
		}
	}
	previous := PowerSupplyBasePath
	PowerSupplyBasePath = base
	t.Cleanup(func() { PowerSupplyBasePath = previous })
}

func TestFindBattery(t *testing.T) {
	fakePowerSupply(t, map[string]map[string]string{
		"BAT0": {"serial_number": "S123", "model_name": "Main"},
		"BAT1": {"serial_number": "S456", "model_name": "Dock"},
		"AC":   {"online": "1"},
	})

	byName, err := FindBattery("bat1")
	assert.NoError(t, err)
	assert.Equal(t, "BAT1", byName.Name)

	bySerial, err := FindBattery("S123")
	assert.NoError(t, err)
	assert.Equal(t, "BAT0", bySerial.Name)

	_, err = FindBattery("nope")
	assert.Error(t, err)

	_, err = FindBattery("")
	assert.Error(t, err)
}

func TestSetChargeControlEndThreshold(t *testing.T) {
	fakePowerSupply(t, map[string]map[string]string{
		"BAT0": {"charge_control_end_threshold": "100"},
	})
	battery, err := FindBattery("BAT0")
	assert.NoError(t, err)

	assert.NoError(t, battery.SetChargeControlEndThreshold(80))
	value, err := battery.ReadChargeControlEndThreshold()
	assert.NoError(t, err)
	assert.Equal(t, 80, value)

	assert.Error(t, battery.SetChargeControlEndThreshold(101))
	assert.Error(t, battery.SetChargeControlEndThreshold(-1))
	value, _ = battery.ReadChargeControlEndThreshold()
	assert.Equal(t, 80, value)
}

func TestGetDegradation(t *testing.T) {
	fakePowerSupply(t, map[string]map[string]string{
		"BAT0": {"energy_full": "45000000", "energy_full_design": "50000000"},
		"BAT1": {"energy_full": "45000000"},
	})

	battery, _ := FindBattery("BAT0")
	degradation, err := battery.GetDegradation()
	assert.NoError(t, err)
	assert.InDelta(t, 10.0, degradation, 0.001)

	missing, _ := FindBattery("BAT1")
	_, err = missing.GetDegradation()
	assert.Error(t, err)
}

func TestCalculateRemainingTime(t *testing.T) {
	assert.Equal(t, int64(3600), CalculateRemainingTime(10, 10))
}
