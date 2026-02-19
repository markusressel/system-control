package device

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get/Set the current profile of a device",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		state := pipewire.PwDump()

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		}

		device, err := state.FindDeviceByName(deviceName)
		if err != nil {
			return err
		}

		if len(profileName) > 0 {
			profile, err := device.GetProfileIdByName(profileName)
			if err != nil {
				return err
			}

			return device.SetProfileByName(profile.Name)
		} else {
			profile, err := device.GetActiveProfile()
			if err != nil {
				return err
			}
			fmt.Println(profile.Description)
			return nil
		}
	},
}

func init() {
	DeviceCmd.AddCommand(profileCmd)
}
