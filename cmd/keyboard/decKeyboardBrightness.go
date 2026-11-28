package keyboard

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var decKeyboardBrightnessCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the keyboard backlight brightness",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("keyboard.brightness.dec"),
}

func init() {
	keyboardBrightnessCmd.AddCommand(decKeyboardBrightnessCmd)
}
