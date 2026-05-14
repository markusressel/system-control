package display

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var displayWakeCmd = &cobra.Command{
	Use:   "wakeup",
	Short: "Wake connected displays",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := util.ExecCommand(
			"xset",
			"dpms",
			"force",
			"on",
		)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Command.AddCommand(displayWakeCmd)
}
