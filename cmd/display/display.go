package display

import (
	"github.com/markusressel/system-control/cmd/display/backlight"
	"github.com/markusressel/system-control/cmd/display/redshift"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "display",
	Short: "Control Displays",
	Long:  ``,
}

func init() {
	Command.AddCommand(backlight.Command)
	Command.AddCommand(redshift.Command)
}
