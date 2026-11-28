package fan

import (
	internalfan "github.com/markusressel/system-control/internal/fan"
	"github.com/spf13/cobra"
)

var modeRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate through available fan modes",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := internalfan.EnsureRoot(); err != nil {
			return err
		}
		return internalfan.RotateMode()
	},
}

func init() {
	modeCmd.AddCommand(modeRotateCmd)
}
