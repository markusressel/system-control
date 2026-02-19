package display

import (
	"fmt"

	"github.com/markusressel/system-control/internal/util"

	"github.com/spf13/cobra"
)

var displayListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current displays",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		displays, err := util.GetDisplays()
		if err != nil {
			return err
		}

		for _, display := range displays {
			fmt.Println(display.Name)
		}

		return nil
	},
}

func init() {
	Command.AddCommand(displayListCmd)
}
