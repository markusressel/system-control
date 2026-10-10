package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftBrightnessResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the display brightness to 1.0.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.brightness.reset"),
}

func init() {
	brightnessCmd.AddCommand(redshiftBrightnessResetCmd)
}
