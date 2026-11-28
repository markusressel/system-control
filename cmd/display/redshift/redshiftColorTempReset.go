package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftColorTempResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the display color temperature to 6500K.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.color-temperature.reset"),
}

func init() {
	colorTempCmd.AddCommand(redshiftColorTempResetCmd)
}
