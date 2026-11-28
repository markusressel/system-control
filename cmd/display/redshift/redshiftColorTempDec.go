package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftColorTempDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift color temperature.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.color-temperature.dec"),
}

func init() {
	redshiftColorTempDecCmd.PersistentFlags().Int64P(
		"step", "s",
		500,
		"Step size to decrease the color temperature by",
	)

	colorTempCmd.AddCommand(redshiftColorTempDecCmd)
}
