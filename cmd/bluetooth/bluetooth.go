package bluetooth

import (
	"github.com/spf13/cobra"
)

var Name string

var Command = &cobra.Command{
	Use:              "bluetooth",
	Short:            "Control Bluetooth Devices",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	Command.PersistentFlags().StringVarP(
		&Name,
		"name", "n",
		"",
		"Device Name",
	)
}
