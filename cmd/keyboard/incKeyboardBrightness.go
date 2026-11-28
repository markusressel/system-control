package keyboard

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var incKeyboardBrightnessCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the keyboard backlight brightness",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("keyboard.brightness.inc"),
}

func init() {
	keyboardBrightnessCmd.AddCommand(incKeyboardBrightnessCmd)
}
