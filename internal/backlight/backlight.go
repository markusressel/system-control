// Package backlight implements the display and keyboard backlight commands.
package backlight

import (
	"fmt"

	"github.com/markusressel/system-control/internal/util"
)

// stepPercent is how many percent one inc/dec changes: the lower the brightness, the finer
// the step, as the eye notices changes in dark rooms much more.
func stepPercent(percentage int) int {
	switch {
	case percentage < 10:
		return 1
	case percentage < 20:
		return 2
	case percentage < 40:
		return 4
	default:
		return 8
	}
}

// rawStep converts the step for the current brightness to raw backlight units. It is at least
// 1 so that a backlight with a small maximum can still be changed.
func rawStep(brightness, maxBrightness int) int {
	percentage := int(float32(brightness) / float32(maxBrightness) * 100.0)
	return max(int(float32(stepPercent(percentage))*(float32(maxBrightness)/100.0)), 1)
}

// displayPercent returns the brightness of the main display backlight in percent.
func displayPercent() (int, error) {
	b, err := util.GetMainBacklight()
	if err != nil {
		return 0, err
	}
	brightness, err := b.GetBrightness()
	if err != nil {
		return 0, err
	}
	maxBrightness, err := b.GetMaxBrightness()
	if err != nil {
		return 0, err
	}
	if maxBrightness <= 0 {
		return 0, fmt.Errorf("backlight %s reports a maximum brightness of %d", b.Name, maxBrightness)
	}
	return int(float32(brightness) / float32(maxBrightness) * 100.0), nil
}

// stepDisplay changes the main display brightness by one step in the given direction (1 or -1).
func stepDisplay(direction int) error {
	b, err := util.GetMainBacklight()
	if err != nil {
		return err
	}
	brightness, err := b.GetBrightness()
	if err != nil {
		return err
	}
	maxBrightness, err := b.GetMaxBrightness()
	if err != nil {
		return err
	}
	if maxBrightness <= 0 {
		return fmt.Errorf("backlight %s reports a maximum brightness of %d", b.Name, maxBrightness)
	}
	return b.AdjustBrightness(direction * rawStep(brightness, maxBrightness))
}

func setDisplay(percentage int) error {
	b, err := util.GetMainBacklight()
	if err != nil {
		return err
	}
	return b.SetBrightness(percentage)
}

func stepKeyboard(direction int) error {
	brightness, err := util.GetKeyboardBrightness()
	if err != nil {
		return err
	}
	_, err = util.SetKeyboardBrightness(brightness + direction)
	return err
}
