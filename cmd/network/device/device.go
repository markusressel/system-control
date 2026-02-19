package device

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "device",
	Short: "Network device specific commands",
	Long:  ``,
}

func init() {
	Command.AddCommand(listCmd)
}
