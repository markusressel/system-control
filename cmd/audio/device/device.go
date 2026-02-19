package device

import (
	"github.com/spf13/cobra"
)

var (
	deviceName string
)

var DeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Show a list of all available devices",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement this
		//result, err := util.ExecCommand("pactl", "list", "sinks")
		//if err != nil {
		//	return err
		//}
		//print(result)
		return nil
	},
}

func init() {
	DeviceCmd.PersistentFlags().StringVarP(
		&deviceName,
		"device", "d",
		"",
		"device",
	)
}
