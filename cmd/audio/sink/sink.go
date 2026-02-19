package sink

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var SinkCmd = &cobra.Command{
	Use:   "sink",
	Short: "Show a list of all available sinks",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := util.ExecCommand("pactl", "list", "sinks")
		if err != nil {
			return err
		}
		print(result)
		return nil
	},
}
