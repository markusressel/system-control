package backlight

import (
	"strconv"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var setBrightnessCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the brightness of a given display backlight.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		mainBacklight, err := util.GetMainBacklight()
		if err != nil {
			return err
		}
		return mainBacklight.SetBrightness(p)
	},
}

func init() {
	brightnessCmd.AddCommand(setBrightnessCmd)
}
