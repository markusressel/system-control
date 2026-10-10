package video

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var loadDriverCmd = &cobra.Command{
	Use:   "load",
	Short: "Load the video driver, allowing access to video devices",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := util.ExecCommand("modprobe", "uvcvideo")
		return err
	},
}

func init() {
	Command.AddCommand(loadDriverCmd)
}
