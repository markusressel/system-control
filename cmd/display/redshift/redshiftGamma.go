package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var gammaCmd = &cobra.Command{
	Use:   "gamma",
	Short: "Show current display gamma",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.gamma"),
}

func init() {
	commandutil.AddWatchFlag(gammaCmd)

	gammaCmd.Flags().Float64P(
		"gamma", "g",
		-1,
		"Gamma value to set (between 0.1 and 2.0)",
	)

	gammaCmd.Flags().StringP(
		"format", "f",
		"%.2f",
		"Format string for the output value",
	)

	Command.AddCommand(gammaCmd)
}
