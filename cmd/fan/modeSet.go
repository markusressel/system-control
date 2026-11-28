package fan

import (
	internalfan "github.com/markusressel/system-control/internal/fan"
	"github.com/spf13/cobra"
)

var modeSetCmd = &cobra.Command{
	Use:   "set [mode]",
	Short: "Set fan mode by number (0=Default, 1=Boost, 2=Silent)",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := internalfan.EnsureRoot(); err != nil {
			return err
		}
		return internalfan.SetModeArg(args[0])
	},
}

func init() {
	modeCmd.AddCommand(modeSetCmd)
}
