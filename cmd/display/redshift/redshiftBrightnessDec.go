package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftBrightnessDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift brightness.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.brightness.dec"),
}

func init() {
	redshiftBrightnessDecCmd.PersistentFlags().Float64P(
		"step", "s",
		0.1,
		"Step size to decrease the brightness by (between 0.1 and 1.0)",
	)

	brightnessCmd.AddCommand(redshiftBrightnessDecCmd)
}
