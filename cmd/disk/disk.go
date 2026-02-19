package disk

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "disk",
	Short: "Display Disk info",
	Long:  ``,
}

func init() {
	Command.AddCommand(diskListCmd)
}
