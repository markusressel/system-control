package cmd

import (
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the configuration of the system-control tool",
	Long:  `Print the configuration of the system-control tool.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configuration.PrintConfig()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
}
