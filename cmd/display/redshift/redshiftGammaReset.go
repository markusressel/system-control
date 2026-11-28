package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var redshiftGammaResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the display gamma to 1.0.",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.gamma.reset"),
}

func init() {
	gammaCmd.AddCommand(redshiftGammaResetCmd)
}
