package battery

import (
	batteryutil "github.com/markusressel/system-control/internal/battery"
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
		batteryutil.NameFlag, "n",
		batteryutil.DefaultName,
		"Battery Name",
	)
}
