package sink

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

const (
	ColumnID          = "id"
	ColumnName        = "name"
	ColumnDescription = "description"
)

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
	RunE: commandutil.OpRunE("audio.sink.active"),
}

func init() {
	activeCmd.Flags().StringSliceP(
		"columns", "c",
		defaultColumns,
		"Columns to print (id,name,description)",
	)
	commandutil.AddWatchFlag(activeCmd)

	SinkCmd.AddCommand(activeCmd)
}
