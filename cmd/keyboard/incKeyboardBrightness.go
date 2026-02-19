package keyboard

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var incKeyboardBrightnessCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the keyboard backlight brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		brightness, _ := util.GetKeyboardBrightness()
		_, err := util.SetKeyboardBrightness(brightness + 1)
		return err
	},
}

func init() {
	keyboardBrightnessCmd.AddCommand(incKeyboardBrightnessCmd)
}
