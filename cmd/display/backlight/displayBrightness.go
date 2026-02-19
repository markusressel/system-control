package backlight

import (
	"fmt"

	"github.com/markusressel/system-control/internal/util"

	"github.com/spf13/cobra"
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Show current display brightness",
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

		fmt.Println(percentage)
		return nil
	},
}

func init() {
	Command.AddCommand(brightnessCmd)
}
