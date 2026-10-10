package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftColorTempIncCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the currently applied redshift color temperature.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.color-temperature.inc"),
}

func init() {
	redshiftColorTempIncCmd.PersistentFlags().Int64P(
		"step", "s",
		500,
		"Step size to increase the color temperature by",
	)

	colorTempCmd.AddCommand(redshiftColorTempIncCmd)
}
