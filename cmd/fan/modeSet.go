package fan

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var modeSetCmd = &cobra.Command{
	Use:   "set [mode]",
	Short: "Set fan mode by number (0=Default, 1=Boost, 2=Silent)",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRoot(); err != nil {
			return err
		}

		mode, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid fan mode %q: %w", args[0], err)
		}

		if mode < fanModeDefault || mode > fanModeSilent {
			return fmt.Errorf("invalid fan mode %d (expected %d, %d, or %d)", mode, fanModeDefault, fanModeBoost, fanModeSilent)
		}

		return setFanMode(mode)
	},
}

func init() {
	modeCmd.AddCommand(modeSetCmd)
}
