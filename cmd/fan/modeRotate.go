package fan

import "github.com/spf13/cobra"

var modeRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate through available fan modes",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRoot(); err != nil {
			return err
		}

		currentMode, err := readFanMode()
		if err != nil {
			return err
		}

		nextMode := (currentMode + 1) % fanModeCount
		return setFanMode(nextMode)
	},
}

func init() {
	modeCmd.AddCommand(modeRotateCmd)
}
