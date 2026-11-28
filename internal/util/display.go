package util

import (
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	xrandrExecutable = "xrandr"
)

// DisplayInfo represents a display
type DisplayInfo struct {
	Name string
}

var displayCache struct {
	mu        sync.Mutex
	ttl       time.Duration
	fetchedAt time.Time
	displays  []DisplayInfo
}

// EnableDisplayCache makes GetDisplays reuse its result for the given duration, so that
// frequent queries of a long running process (the daemon) do not run xrandr each time.
// A duration <= 0 disables the cache again.
func EnableDisplayCache(ttl time.Duration) {
	displayCache.mu.Lock()
	defer displayCache.mu.Unlock()
	displayCache.ttl = ttl
	displayCache.displays = nil
}

// GetDisplays returns a list of all connected displays
func GetDisplays() (displays []DisplayInfo, err error) {
	displayCache.mu.Lock()
	defer displayCache.mu.Unlock()

	if displayCache.ttl > 0 && displayCache.displays != nil && time.Since(displayCache.fetchedAt) < displayCache.ttl {
		return slices.Clone(displayCache.displays), nil
	}

	displays, err = queryDisplays()
	if err == nil && displayCache.ttl > 0 {
		displayCache.displays = slices.Clone(displays)
		if displayCache.displays == nil {
			displayCache.displays = []DisplayInfo{}
		}
		displayCache.fetchedAt = time.Now()
	}
	return displays, err
}

func queryDisplays() (displays []DisplayInfo, err error) {
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

// DisplayConfig represents a display configuration
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

// SetDisplayConfig sets the display configuration for a single display.
func SetDisplayConfig(displayConfig DisplayConfig) error {
	return SetDisplayConfigs([]DisplayConfig{displayConfig})
}

// SetDisplayConfigs sets the display configuration for multiple displays.
func SetDisplayConfigs(displayConfigs []DisplayConfig) error {
	_, err := ExecCommand(xrandrExecutable, XrandrArgs(displayConfigs)...)
	return err
}

// XrandrArgs builds the xrandr arguments that apply the given display configurations.
func XrandrArgs(displayConfigs []DisplayConfig) []string {
	args := []string{}

	for _, displayConfig := range displayConfigs {
		args = append(args, "--output", displayConfig.Name)

		if displayConfig.Off {
			// all other options are meaningless for a disabled display
			args = append(args, "--off")
			continue
		}

		if displayConfig.Primary {
			args = append(args, "--primary")
		}
		if displayConfig.Auto {
			args = append(args, "--auto")
		} else if IsNotEmpty(displayConfig.Mode) {
			args = append(args, "--mode", displayConfig.Mode)
		}
		if IsNotEmpty(displayConfig.Position) {
			args = append(args, "--pos", displayConfig.Position)
		}
		if displayConfig.Rate > 0 {
			args = append(args, "--rate", strconv.Itoa(displayConfig.Rate))
		}
	}

	return args
}
