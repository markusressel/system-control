package cmd

import (
	"github.com/spf13/cobra"
)

var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown the System gracefully",
	Long:  `Shuts down the system in a graceful way, first closing all opened applications.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return gracefulPowerAction("poweroff")
	},
}

func init() {
	RootCmd.AddCommand(shutdownCmd)
	addPowerActionFlags(shutdownCmd)
}
