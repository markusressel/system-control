package disk

import (
	"fmt"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var diskListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current disks",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		disks, err := util.GetDisks()
		if err != nil {
			return err
		}

		for _, disk := range disks {
			fmt.Println(disk.Name)
			fetchAndPrintSmartCtlData(disk)
		}

		return nil
	},
}

// fetchAndPrintSmartCtlData prints the SMART data in a formatted way
func fetchAndPrintSmartCtlData(disk util.DiskInfo) {
	smartCtlData, err := disk.GetSmartCtlData()
	if err != nil {
		return
	}
	properties := orderedmap.NewOrderedMap[string, string]()
	properties.Set("Temperature", fmt.Sprintf("%d", smartCtlData.Temperature.Current))

	// check if any attribute is below threshold, if so, print them as well
	for _, attr := range smartCtlData.AtaSmartAttributes.Table {
		if attr.WhenFailed != "" || attr.Value <= attr.Thresh {
			properties.Set(attr.Name, fmt.Sprintf("Raw: %d, Value: %d, Worst: %d, Thresh: %d, WhenFailed: %s", attr.Raw.Value, attr.Value, attr.Worst, attr.Thresh, attr.WhenFailed))
		}
	}

	util.PrintFormattedTableOrdered(fmt.Sprintf("SmartCtl"), properties)

}

func init() {
	Command.AddCommand(diskListCmd)
}
