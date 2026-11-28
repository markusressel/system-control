package display

import (
	"fmt"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

// The per display options are matched to --display by position.
var (
	displays  []string
	modes     []string
	positions []string
	rate      []int
	primary   []bool
	off       []bool
	auto      []bool
)

var displayConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Set the desired display configuration",
	Long: `Set the configuration of one or more displays.

All per display options are matched to --display by position, e.g.
  display config -d DisplayPort-1 -d DisplayPort-2 -m 3840x2160 -m 2560x1440`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		displayConfigs, err := createDisplayConfigurationsFromArgs(displays, modes, positions, rate, primary, off, auto)
		if err != nil {
			return err
		}
		return util.SetDisplayConfigs(displayConfigs)
	},
}

func createDisplayConfigurationsFromArgs(
	displayNameArgs []string,
	modeArgs []string,
	positionArgs []string,
	rateArgs []int,
	primaryArgs []bool,
	offArgs []bool,
	autoArgs []bool,
) ([]util.DisplayConfig, error) {
	options := map[string]int{
		"mode":     len(modeArgs),
		"position": len(positionArgs),
		"rate":     len(rateArgs),
		"primary":  len(primaryArgs),
		"off":      len(offArgs),
		"auto":     len(autoArgs),
	}
	for name, count := range options {
		if count > len(displayNameArgs) {
			return nil, fmt.Errorf("got %d --%s values for %d displays", count, name, len(displayNameArgs))
		}
	}

	displayConfigs := make([]util.DisplayConfig, 0, len(displayNameArgs))
	for i, displayName := range displayNameArgs {
		displayConfig := util.NewDisplayConfig(displayName)
		if i < len(modeArgs) {
			displayConfig.Mode = modeArgs[i]
		}
		if i < len(positionArgs) {
			displayConfig.Position = positionArgs[i]
		}
		if i < len(rateArgs) {
			displayConfig.Rate = rateArgs[i]
		}
		if i < len(primaryArgs) {
			displayConfig.Primary = primaryArgs[i]
		}
		if i < len(offArgs) {
			displayConfig.Off = offArgs[i]
		}
		if i < len(autoArgs) {
			displayConfig.Auto = autoArgs[i]
		}
		displayConfigs = append(displayConfigs, displayConfig)
	}

	return displayConfigs, nil
}

func init() {
	Command.AddCommand(displayConfigCmd)

	displayConfigCmd.Flags().StringSliceVarP(&displays, "display", "d", []string{}, "The display to configure")
	_ = displayConfigCmd.MarkFlagRequired("display")

	displayConfigCmd.Flags().StringSliceVarP(&modes, "mode", "m", []string{}, "The display mode to set")
	displayConfigCmd.Flags().StringSliceVarP(&positions, "position", "p", []string{}, "The display position to set")
	displayConfigCmd.Flags().IntSliceVarP(&rate, "rate", "r", []int{}, "The display refresh rate to set")
	displayConfigCmd.Flags().BoolSliceVarP(&primary, "primary", "P", []bool{}, "Set the display as primary")
	displayConfigCmd.Flags().BoolSliceVarP(&off, "off", "O", []bool{}, "Turn the display off")
	displayConfigCmd.Flags().BoolSliceVarP(&auto, "auto", "A", []bool{}, "Set the display mode automatically")
}
