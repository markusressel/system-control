package device

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get/Set the current profile of a device",
	Long:  ``,
	RunE:  commandutil.OpRunE("audio.device.profile"),
}

func init() {
	commandutil.AddWatchFlag(profileCmd)

	DeviceCmd.AddCommand(profileCmd)
}
