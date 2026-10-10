package backlight

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var incBrightnessCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase display backlight brightness",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("display.backlight.brightness.inc"),
}

func init() {
	brightnessCmd.AddCommand(incBrightnessCmd)
}
