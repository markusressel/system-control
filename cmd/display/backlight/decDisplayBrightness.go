package backlight

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var decBrightnessCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease display backlight brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		mainBacklight, err := util.GetMainBacklight()
		if err != nil {
			return err
		}
		brightness, err := mainBacklight.GetBrightness()
		if err != nil {
			return err
		}
		maxBrightness, err := mainBacklight.GetMaxBrightness()
		if err != nil {
			return err
		}

		percentage := int((float32(brightness) / float32(maxBrightness)) * 100.0)

		var change int
		if percentage < 10 {
			change = 1
		} else if percentage < 20 {
			change = 2
		} else if percentage < 40 {
			change = 4
		} else {
			change = 8
		}

		rawChange := int(float32(change) * (float32(maxBrightness) / 100.0))

		return mainBacklight.AdjustBrightness(-rawChange)
	},
}

func init() {
	brightnessCmd.AddCommand(decBrightnessCmd)
}
