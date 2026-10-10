package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftGammaDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift gamma.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.gamma.dec"),
}

func init() {
	redshiftGammaDecCmd.PersistentFlags().Float64P(
		"step", "s",
		0.1,
		"Step size to decrease the gamma by",
	)

	gammaCmd.AddCommand(redshiftGammaDecCmd)
}
