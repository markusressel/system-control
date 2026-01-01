package disk

import (
	"fmt"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var diskListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current disks",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		disks, err := util.GetDisks()
		if err != nil {
			return err
		}

		for _, disk := range disks {
			fmt.Println(disk.Name)
		}

		return nil
	},
}

func init() {
	Command.AddCommand(diskListCmd)
}
