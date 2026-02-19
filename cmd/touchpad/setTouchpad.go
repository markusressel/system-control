package touchpad

import (
	"strconv"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var setTouchpadCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the enabled state of the Touchpad",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		enabled, err := strconv.ParseBool(args[0])
		if err != nil {
			return err
		}

		return util.SetTouchpadEnabled(enabled)
	},
}

func init() {
	Command.AddCommand(setTouchpadCmd)
}
