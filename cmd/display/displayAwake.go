package display

import (
	"fmt"
	"strings"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var displayAwakeCmd = &cobra.Command{
	Use:   "awake",
	Short: "Check whether displays are awake",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := util.ExecCommand(
			"xset",
			"dpms",
			"q",
		)
		if err != nil {
			return err
		}

		switch {
		case strings.Contains(output, "Monitor is On"):
			fmt.Println("yes")
			return nil
		case strings.Contains(output, "Monitor is Off"):
			fmt.Println("no")
			return nil
		default:
			return fmt.Errorf("unable to determine monitor state from xset output")
		}
	},
}

func init() {
	Command.AddCommand(displayAwakeCmd)
}
