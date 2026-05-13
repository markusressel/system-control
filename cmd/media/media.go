package media

import (
	"fmt"

	internalmedia "github.com/markusressel/system-control/internal/media"
	"github.com/spf13/cobra"
)

var player string

var Command = &cobra.Command{
	Use:              "media",
	Short:            "Control media players via playerctl",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("status", true)
	},
}

func printPlayerCtlOutput(command string, targetAllWhenUnspecified bool) error {
	output, err := internalmedia.RunPlayerCtl(command, player, targetAllWhenUnspecified)
	if err != nil {
		return err
	}
	if output != "" {
		fmt.Println(output)
	}
	return nil
}

func init() {
	Command.PersistentFlags().StringVarP(
		&player,
		"player", "p",
		"",
		"Player name or regex",
	)
}
