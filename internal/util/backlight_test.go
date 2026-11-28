package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fakeBacklight(t *testing.T, name string, brightness string, max string) {
	base := t.TempDir()
	dir := filepath.Join(base, name)
	assert.NoError(t, os.MkdirAll(dir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, Brightness), []byte(brightness), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, MaxBrightness), []byte(max), 0644))

	previous := DisplayBacklightPath
	DisplayBacklightPath = base
	t.Cleanup(func() { DisplayBacklightPath = previous })
}

func TestBacklightAdjustIsClamped(t *testing.T) {
	fakeBacklight(t, "intel_backlight", "90", "100")
	backlight, err := GetMainBacklight()
	assert.NoError(t, err)

	assert.NoError(t, backlight.AdjustBrightness(50))
	value, _ := backlight.GetBrightness()
	assert.Equal(t, 100, value)

	assert.NoError(t, backlight.AdjustBrightness(-500))
	value, _ = backlight.GetBrightness()
	assert.Equal(t, 0, value)
}

func TestBacklightSetPercentageIsClamped(t *testing.T) {
	fakeBacklight(t, "intel_backlight", "10", "200")
	backlight, _ := GetMainBacklight()

	assert.NoError(t, backlight.SetBrightness(50))
	value, _ := backlight.GetBrightness()
	assert.Equal(t, 100, value)

	assert.NoError(t, backlight.SetBrightness(500))
	value, _ = backlight.GetBrightness()
	assert.Equal(t, 200, value)
}

func TestNoBacklight(t *testing.T) {
	previous := DisplayBacklightPath
	DisplayBacklightPath = t.TempDir()
	t.Cleanup(func() { DisplayBacklightPath = previous })

	_, err := GetMainBacklight()
	assert.Error(t, err)
}
