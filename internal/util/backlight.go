package util

import (
	"errors"
	"fmt"
	"os"
)

// GetBacklights returns all found backlights
func GetBacklights() ([]Backlight, error) {
	files, err := os.ReadDir(DisplayBacklightPath)
	if err != nil {
		return nil, err
	}

	var backlights []Backlight
	for _, file := range files {
		backlights = append(backlights, NewBacklight(file.Name()))
	}

	return backlights, nil
}

// GetMainBacklight returns the first found backlight
func GetMainBacklight() (Backlight, error) {
	backlights, err := GetBacklights()
	if err != nil {
		return Backlight{}, err
	}
	if len(backlights) == 0 {
		return Backlight{}, errors.New("no backlights found")
	}
	mainBacklight := backlights[0]
	if len(backlights) > 1 {
		fmt.Printf("multiple backlights found, using first as main: %s\n", mainBacklight.Name)
	}

	return mainBacklight, nil
}

// Represents a display backlight
type Backlight struct {
	Name string

	brightnessPath    string
	maxBrightnessPath string
}

func NewBacklight(name string) Backlight {
	return Backlight{
		Name:              name,
		brightnessPath:    computeBacklightPropertyPath(name, Brightness),
		maxBrightnessPath: computeBacklightPropertyPath(name, MaxBrightness),
	}
}

func (b Backlight) GetBrightness() (int, error) {
	brightnessPath := b.brightnessPath
	brightness, err := ReadIntFromFile(brightnessPath)
	if err != nil {
		return -1, err
	}
	return int(brightness), nil
}

// SetBrightness sets the brightness of the main display to the given percentage
// Note: This function only works if the display has a max_brightness value
func (b Backlight) SetBrightness(percentage int) error {
	maxBrightnessPath := b.maxBrightnessPath
	brightnessPath := b.brightnessPath

	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		return err
	}

	targetValue := int((float32(percentage) / 100.0) * float32(maxBrightness))
	return WriteIntToFile(targetValue, brightnessPath)
}

// AdjustBrightness adjusts the brightness of the main display
func (b Backlight) AdjustBrightness(change int) error {
	maxBrightness, err := b.GetMaxBrightness()
	if err != nil {
		return err
	}
	currentBrightness, err := b.GetBrightness()
	if err != nil {
		return err
	}

	targetBrightness := currentBrightness + change
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	return b.setBrightnessRaw(targetBrightness)
}

func (b Backlight) GetMaxBrightness() (int, error) {
	maxBrightnessPath := b.maxBrightnessPath
	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		return -1, err
	}
	return int(maxBrightness), nil
}

// setBrightnessRaw sets the brightness of the given display backlight to the given value
func (b Backlight) setBrightnessRaw(brightness int) error {
	maxBrightness, err := b.GetMaxBrightness()
	if err != nil {
		return err
	}
	targetBrightness := brightness
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	return WriteIntToFile(targetBrightness, b.brightnessPath)
}

func computeBacklightPropertyPath(backlight string, property string) string {
	return DisplayBacklightPath + string(os.PathSeparator) + backlight + string(os.PathSeparator) + property
}
