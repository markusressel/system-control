package fan

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	fanModeTypeFlag        = "type"
	fanModeTypeText        = "text"
	fanModeTypeNumber      = "number"
	fanPolicyFileGlob      = "/sys/devices/platform/asus-nb-wmi/hwmon/*/device/throttle_thermal_policy"
	intelPstateNoTurboFile = "/sys/devices/system/cpu/intel_pstate/no_turbo"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		modeType, err := cmd.Flags().GetString(fanModeTypeFlag)
		if err != nil {
			return err
		}

		mode, err := readFanMode()
		if err != nil {
			return err
		}

		switch modeType {
		case fanModeTypeNumber:
			fmt.Println(mode)
			return nil
		case fanModeTypeText:
			modeName, ok := fanModeNames[mode]
			if !ok {
				fmt.Printf("Unknown (%d)\n", mode)
				return nil
			}
			fmt.Println(modeName)
			return nil
		default:
			return fmt.Errorf("invalid value for --%s: %q (expected %q or %q)", fanModeTypeFlag, modeType, fanModeTypeText, fanModeTypeNumber)
		}
	},
}

func init() {
	modeCmd.Flags().String(fanModeTypeFlag, fanModeTypeText, "output format: text or number")
}

func ensureRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("please run as root")
	}

	return nil
}

func readFanMode() (int, error) {
	policyFiles, err := filepath.Glob(fanPolicyFileGlob)
	if err != nil {
		return 0, err
	}
	if len(policyFiles) == 0 {
		return 0, fmt.Errorf("no fan policy files found matching %q", fanPolicyFileGlob)
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
	policyFiles, err := filepath.Glob(fanPolicyFileGlob)
	if err != nil {
		return err
	}
	if len(policyFiles) == 0 {
		return fmt.Errorf("no fan policy files found matching %q", fanPolicyFileGlob)
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

	if err := os.WriteFile(intelPstateNoTurboFile, []byte(noTurbo), 0); err != nil {
		return err
	}

	return nil
}
