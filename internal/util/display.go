package util

import (
	"strconv"
	"strings"
)

const (
	xrandrExecutable = "xrandr"
)

type DisplayInfo struct {
	Name string
}

func GetDisplays() (displays []DisplayInfo, err error) {
	result, err := ExecCommand(
		xrandrExecutable,
		"--listmonitors",
	)
	if err != nil {
		return nil, err
	}

	resultLines := strings.Split(result, "\n")

	for i, line := range resultLines {
		if i == 0 {
			continue
		} else {
			segments := strings.Split(line, " ")
			displayName := segments[len(segments)-1]
			displays = append(displays, DisplayInfo{
				Name: displayName,
			})
		}
	}
	return displays, nil
}

type DisplayConfig struct {
	// Name is the name of the display as shown by GetDisplays
	Name string

	// Off is a boolean that represents whether the display should be turned off.
	// Note: If Off is set to true, all other options will be ignored.
	Off bool

	// Primary is a boolean that represents whether the display is the primary display.
	// Note: Only one display configuration can be primary.
	Primary bool

	// Auto is a boolean that represents whether the display mode should be set automatically.
	// Note: If Auto is set to true, Mode (and Position?) will be ignored.
	Auto bool

	// Mode is a string that represents the desired display mode.
	// Example: 3840x2160
	Mode string

	// Position is a string that represents the positioning of the screen content within the total virtual screen space.
	// Example: 3840x0
	Position string

	// Rate is an integer that represents the desired refresh rate.
	Rate int
}

func NewDisplayConfig(name string) DisplayConfig {
	return DisplayConfig{
		Name: name,
	}
}

func (displayConfig DisplayConfig) SetOff(off bool) DisplayConfig {
	if off {
		displayConfig.SetPrimary(false)
		displayConfig.SetAuto(false)
		displayConfig.Mode = ""
		displayConfig.Position = ""
	}
	displayConfig.Off = off
	return displayConfig
}

func (displayConfig DisplayConfig) SetPrimary(primary bool) DisplayConfig {
	if primary {
		displayConfig.SetOff(false)
	}
	displayConfig.Primary = primary
	return displayConfig
}

func (displayConfig DisplayConfig) SetAuto(auto bool) DisplayConfig {
	if auto {
		displayConfig.SetOff(false)
		displayConfig.Mode = ""
		displayConfig.Position = ""
	}
	displayConfig.Auto = auto
	return displayConfig
}

func (displayConfig DisplayConfig) SetMode(mode string) DisplayConfig {
	if IsNotEmpty(mode) {
		displayConfig.SetOff(false)
		displayConfig.SetAuto(false)
	}
	displayConfig.Mode = mode
	return displayConfig
}

func (displayConfig DisplayConfig) SetPosition(position string) DisplayConfig {
	if IsNotEmpty(position) {
		displayConfig.SetOff(false)
	}
	displayConfig.Position = position
	return displayConfig
}

// SetDisplayConfig sets the display configuration for a single display.
func SetDisplayConfig(displayConfig DisplayConfig) error {
	return SetDisplayConfigs([]DisplayConfig{displayConfig})
}

// SetDisplayConfigs sets the display configuration for multiple displays.
func SetDisplayConfigs(displayConfigs []DisplayConfig) error {
	args := []string{}

	for _, displayConfig := range displayConfigs {
		args = append(args, "--output")
		args = append(args, displayConfig.Name)

		if displayConfig.Off {
			args = append(args, "--off")
		}
		if displayConfig.Primary {
			args = append(args, "--primary")
		}
		if displayConfig.Auto {
			args = append(args, "--auto")
		}
		if IsNotEmpty(displayConfig.Mode) {
			args = append(args, "--mode")
			args = append(args, displayConfig.Mode)
		}
		if IsNotEmpty(displayConfig.Position) {
			args = append(args, "--pos")
			args = append(args, displayConfig.Position)
		}
		if displayConfig.Rate > 0 {
			args = append(args, "--rate")
			args = append(args, strconv.Itoa(displayConfig.Rate))
		}
	}

	_, err := ExecCommand(xrandrExecutable, args...)
	if err != nil {
		return err
	}
	return nil
}
