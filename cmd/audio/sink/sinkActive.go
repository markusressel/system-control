package sink

import (
	"fmt"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

const (
	ColumnID          = "id"
	ColumnName        = "name"
	ColumnDescription = "description"
)

var columns []string
var defaultColumns = []string{ColumnID, ColumnName, ColumnDescription}

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

			for _, col := range columns {
				switch col {
				case ColumnID:
					fmt.Println(node.Id)
				case ColumnName:
					name, err := node.GetName()
					if err != nil {
						return err
					}
					fmt.Println(name)
				case ColumnDescription:
					desc, err := node.GetDescription()
					if err != nil {
						return err
					}
					fmt.Println(desc)
				default:
					return fmt.Errorf("unknown column: %s", col)
				}
			}
		}

		return nil
	},
}

func init() {
	activeCmd.Flags().StringSliceVarP(
		&columns,
		"columns", "c",
		defaultColumns,
		"Columns to print (id,name,description)",
	)

	SinkCmd.AddCommand(activeCmd)
}
