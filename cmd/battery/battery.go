package battery

import (
	"github.com/spf13/cobra"
)

var Name string

var Command = &cobra.Command{
	Use:              "battery",
	Short:            "Control System Battery",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	Command.PersistentFlags().StringVarP(
		&Name,
		"name", "n",
		"BAT0",
		"Battery Name",
	)
}
