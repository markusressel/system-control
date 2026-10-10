package cmd

import (
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Reboot the System gracefully",
	Long:  `Reboots the system gracefully by first closing all currently open windows.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return gracefulPowerAction("reboot")
	},
}

func init() {
	RootCmd.AddCommand(restartCmd)
	addPowerActionFlags(restartCmd)
}
