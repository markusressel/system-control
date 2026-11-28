package backlight

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var setBrightnessCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the brightness of a given display backlight.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  commandutil.OpRunE("display.backlight.brightness.set"),
}

func init() {
	brightnessCmd.AddCommand(setBrightnessCmd)
}
