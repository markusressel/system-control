package sink

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Get active sink index",
	Long: `Get the index of the currently active sink, or check if a given text is part of the active sink:

> system-control audio sink active "headphone"
1

> system-control audio sink active
3`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		searchString := ""
		if len(args) > 0 {
			searchString = args[0]
		}

		state := pipewire.PwDump()

		if len(searchString) > 0 {
			fmt.Println(state.ContainsActiveSink(searchString))
		} else {
			node, err := state.GetDefaultSinkNode()
			if err != nil {
				return err
			}
			fmt.Println(node.Id)
			name, err := node.GetName()
			description, err := node.GetDescription()
			if err == nil {
				fmt.Println(name)
				fmt.Println(description)
			}
		}

		return nil
	},
}

func init() {
	SinkCmd.AddCommand(activeCmd)
}
