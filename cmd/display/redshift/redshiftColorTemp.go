package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var colorTempCmd = &cobra.Command{
	Use:   "color-temperature",
	Short: "Show current display color temperature",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.color-temperature"),
}

func init() {
	commandutil.AddWatchFlag(colorTempCmd)

	colorTempCmd.Flags().Int64P(
		"temperature", "t",
		-1,
		"Color temperature to apply (between 1000 and 25000)")

	Command.AddCommand(colorTempCmd)
}
