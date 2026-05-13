package media

import (
	"fmt"

	internalmedia "github.com/markusressel/system-control/internal/media"
	"github.com/spf13/cobra"
)

var playersCmd = &cobra.Command{
	Use:   "players",
	Short: "List available players",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		players, err := internalmedia.ListPlayers()
		if err != nil {
			return err
		}

		matches, err := internalmedia.MatchPlayers(players, player)
		if err != nil {
			return err
		}

		for _, match := range matches {
			fmt.Println(match)
		}
		return nil
	},
}

func init() {
	Command.AddCommand(playersCmd)
}
