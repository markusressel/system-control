package touchpad

import (
	"github.com/spf13/cobra"

	// registers the ops the commands run
	_ "github.com/markusressel/system-control/internal/touchpad"
)

var Command = &cobra.Command{
	Use:   "touchpad",
	Short: "Control Touchpads",
	Long:  ``,
}
