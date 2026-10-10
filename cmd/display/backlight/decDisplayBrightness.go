package backlight

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var decBrightnessCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease display backlight brightness",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("display.backlight.brightness.dec"),
}

func init() {
	brightnessCmd.AddCommand(decBrightnessCmd)
}
