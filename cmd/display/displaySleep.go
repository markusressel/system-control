package display

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var displaySleepCmd = &cobra.Command{
	Use:   "sleep",
	Short: "Put connected displays to sleep",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := util.ExecCommand(
			"xset",
			"dpms",
			"force",
			"off",
		)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Command.AddCommand(displaySleepCmd)
}
