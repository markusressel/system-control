package keyboard

import (
	"github.com/markusressel/system-control/internal/commandutil"

	"github.com/spf13/cobra"
)

var keyboardBrightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Get/Set the current keyboard backlight brightness",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE:  commandutil.OpRunE("keyboard.brightness"),
}

func init() {
	keyboardBrightnessCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
	Command.AddCommand(keyboardBrightnessCmd)
}
