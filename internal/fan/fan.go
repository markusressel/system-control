package fan

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

const (
	// TypeFlag is the name of the flag that selects the output format of `fan mode`.
	TypeFlag = "type"
	// TypeText prints the name of the fan mode.
	TypeText = "text"
	// TypeNumber prints the number of the fan mode.
	TypeNumber = "number"
	// defaultFanPolicyFileGlob is the fan policy of ASUS laptops, see fan.policyFileGlob in the configuration
	defaultFanPolicyFileGlob = "/sys/devices/platform/asus-nb-wmi/hwmon/*/device/throttle_thermal_policy"
	intelPstateNoTurboFile   = "/sys/devices/system/cpu/intel_pstate/no_turbo"

	// ModeDefault is the default fan mode.
	ModeDefault = 0
	// ModeBoost is the fan mode for maximum performance.
	ModeBoost = 1
	// ModeSilent is the quiet fan mode, it also disables the turbo of Intel CPUs.
	ModeSilent = 2
	// ModeCount is the number of fan modes.
	ModeCount = 3
)

var modeNames = map[int]string{
	ModeDefault: "Default",
	ModeBoost:   "Boost",
	ModeSilent:  "Silent",
}

// EnsureRoot returns an error unless the process runs as root, which writing the fan mode needs.
func EnsureRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("please run as root")
	}

	return nil
}

func fanPolicyFiles() ([]string, error) {
	pattern := configuration.Get().Fan.PolicyFileGlob
	if pattern == "" {
		pattern = defaultFanPolicyFileGlob
	}
	policyFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(policyFiles) == 0 {
		return nil, fmt.Errorf("no fan policy files found matching %q", pattern)
	}
	return policyFiles, nil
}

// ReadMode returns the current fan mode.
func ReadMode() (int, error) {
	policyFiles, err := fanPolicyFiles()
	if err != nil {
		return 0, err
	}

	fileContent, err := os.ReadFile(policyFiles[0])
	if err != nil {
		return 0, err
	}

	mode, err := strconv.Atoi(strings.TrimSpace(string(fileContent)))
	if err != nil {
		return 0, err
	}

	return mode, nil
}

// SetMode writes the fan mode to every policy file and switches the Intel turbo off in silent mode.
func SetMode(mode int) error {
	policyFiles, err := fanPolicyFiles()
	if err != nil {
		return err
	}

	modeString := strconv.Itoa(mode)
	for _, policyFile := range policyFiles {
		if err := os.WriteFile(policyFile, []byte(modeString), 0); err != nil {
			return err
		}
	}

	noTurbo := "0"
	if mode == ModeSilent {
		noTurbo = "1"
	}

	// not available on every CPU (e.g. AMD)
	if _, err := os.Stat(intelPstateNoTurboFile); err == nil {
		if err := os.WriteFile(intelPstateNoTurboFile, []byte(noTurbo), 0); err != nil {
			return err
		}
	}

	daemon.Notify("fan.")
	return nil
}

// RotateMode switches to the next fan mode.
func RotateMode() error {
	currentMode, err := ReadMode()
	if err != nil {
		return err
	}

	nextMode := (currentMode + 1) % ModeCount
	return SetMode(nextMode)
}

// SetModeArg parses and validates a fan mode given on the command line and sets it.
func SetModeArg(arg string) error {
	mode, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid fan mode %q: %w", arg, err)
	}

	if mode < ModeDefault || mode > ModeSilent {
		return fmt.Errorf("invalid fan mode %d (expected %d, %d, or %d)", mode, ModeDefault, ModeBoost, ModeSilent)
	}

	return SetMode(mode)
}
