package keyboard

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var incKeyboardBrightnessCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the keyboard backlight brightness",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		brightness, err := util.GetKeyboardBrightness()
		if err != nil {
			return err
		}
		_, err = util.SetKeyboardBrightness(brightness + 1)
		return err
	},
}

func init() {
	keyboardBrightnessCmd.AddCommand(incKeyboardBrightnessCmd)
}
