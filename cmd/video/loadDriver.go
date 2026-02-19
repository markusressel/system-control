package video

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var loadDriverCmd = &cobra.Command{
	Use:   "load",
	Short: "Load the video driver, allowing access to video devices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		util.ExecCommand("modprobe", "uvcvideo")
	},
}

func init() {
	Command.AddCommand(loadDriverCmd)
}
