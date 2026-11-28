package backlight

import (
	"github.com/markusressel/system-control/internal/commandutil"

	"github.com/spf13/cobra"
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Show current display brightness",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("display.backlight.brightness"),
}

func init() {
	commandutil.AddWatchFlag(brightnessCmd)
	Command.AddCommand(brightnessCmd)
}
