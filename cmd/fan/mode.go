package fan

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
)

const (
	fanModeTypeFlag   = "type"
	fanModeTypeText   = "text"
	fanModeTypeNumber = "number"
	// defaultFanPolicyFileGlob is the fan policy of ASUS laptops, see fan.policyFileGlob in the configuration
	defaultFanPolicyFileGlob = "/sys/devices/platform/asus-nb-wmi/hwmon/*/device/throttle_thermal_policy"
	intelPstateNoTurboFile   = "/sys/devices/system/cpu/intel_pstate/no_turbo"

	fanModeDefault = 0
	fanModeBoost   = 1
	fanModeSilent  = 2
	fanModeCount   = 3
)

var fanModeNames = map[int]string{
	fanModeDefault: "Default",
	fanModeBoost:   "Boost",
	fanModeSilent:  "Silent",
}

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Get the current fan mode",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("fan.mode"),
}

func init() {
	modeCmd.Flags().String(fanModeTypeFlag, fanModeTypeText, "output format: text or number")
	modeCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
}

func ensureRoot() error {
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

func readFanMode() (int, error) {
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

func setFanMode(mode int) error {
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
	if mode == fanModeSilent {
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
