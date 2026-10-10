package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftGammaIncCmd = &cobra.Command{
	Use:   "inc",
	Short: "Increase the currently applied redshift gamma.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.gamma.inc"),
}

func init() {
	redshiftGammaIncCmd.PersistentFlags().Float64P(
		"step", "s",
		0.1,
		"Step size to increase the gamma by",
	)

	gammaCmd.AddCommand(redshiftGammaIncCmd)
}
