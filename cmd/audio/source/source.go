package source

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var SourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Show a list of all available sources (inputs, f.e. microphones)",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := util.ExecCommand("pactl", "list", "sources")
		if err != nil {
			return err
		}
		print(result)
		return nil
	},
}
