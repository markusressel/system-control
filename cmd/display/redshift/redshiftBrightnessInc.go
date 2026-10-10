package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftBrightnessIncCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the currently applied redshift brightness.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.brightness.inc"),
}

func init() {
	redshiftBrightnessIncCmd.PersistentFlags().Float64VarP(
		&stepFloat,
		"step", "s",
		0.1,
		"Step size to increase the brightness by (between 0.1 and 1.0)",
	)

	brightnessCmd.AddCommand(redshiftBrightnessIncCmd)
}
