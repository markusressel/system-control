package cmd

import (
	"strings"
	"time"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Reboot the System gracefully",
	Long:  `Reboots the system gracefully by first closing all currently open windows.`,
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

		_, err = util.ExecCommand("reboot")
		return err
	},
}

func init() {
	RootCmd.AddCommand(restartCmd)
}
