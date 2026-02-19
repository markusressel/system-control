package video

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var unloadDriverCmd = &cobra.Command{
	Use:   "unload",
	Short: "Unload the video driver, preventing access to any video device",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := util.ExecCommand("rmmod", "-f", "uvcvideo")
		return err
	},
}

func init() {
	Command.AddCommand(unloadDriverCmd)
}
