package cmd

import (
	"strings"
	"time"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown the System gracefully",
	Long:  `Shuts down the system in a graceful way, first closing all opened applications.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		openWindows, err := util.FindOpenWindows()
		if err != nil {
			return err
		}

		for _, element := range openWindows {
			windowId := strings.Split(element, " ")[0]
			_, err := util.ExecCommand("wmctrl", "-i", "-c", windowId)
			if err != nil {
				return err
			}
		}

		// wait for all windows to disappear
		for {
			openWindows, err = util.FindOpenWindows()
			if err != nil {
				return err
			}
			if len(openWindows) <= 0 {
				break
			} else {
				time.Sleep(time.Second)
			}
		}

		_, err = util.ExecCommand("poweroff")
		return err
	},
}

func init() {
	RootCmd.AddCommand(shutdownCmd)
}
