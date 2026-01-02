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
		var printAllAttributes = true

		disks, err := util.GetDisks()
		if err != nil {
			return err
		}

		for _, disk := range disks {
			fmt.Println(disk.Name)
			fetchAndPrintSmartCtlData(disk, printAllAttributes)
		}

		return nil
	},
}

// fetchAndPrintSmartCtlData prints the SMART data in a formatted way
func fetchAndPrintSmartCtlData(disk util.DiskInfo, printAllAttributes bool) {
	smartCtlData, err := disk.GetSmartCtlData()
	if err != nil {
		fmt.Printf("  Error fetching SMART data: %v\n", err)
		return
	}
	properties := orderedmap.NewOrderedMap[string, string]()
	properties.Set("Temperature", fmt.Sprintf("%d", smartCtlData.Temperature.Current))

	// check if any attribute is below threshold, if so, print them as well
	for _, attr := range smartCtlData.AtaSmartAttributes.Table {
		if printAllAttributes || attr.WhenFailed != "" || attr.Value <= attr.Thresh {
			properties.Set(attr.Name, fmt.Sprintf("Raw: %d, Value: %d, Worst: %d, Thresh: %d, WhenFailed: %s", attr.Raw.Value, attr.Value, attr.Worst, attr.Thresh, attr.WhenFailed))
		}
	}

	// also check nvme log info
	if smartCtlData.NvmeSmartHealthInformationLog.AvailableSpare < smartCtlData.NvmeSmartHealthInformationLog.AvailableSpareThreshold {
		properties.Set("Available Spare", fmt.Sprintf("Raw: %d, Threshold: %d", smartCtlData.NvmeSmartHealthInformationLog.AvailableSpare, smartCtlData.NvmeSmartHealthInformationLog.AvailableSpareThreshold))
	}
	if smartCtlData.NvmeSmartHealthInformationLog.PercentageUsed > 90 {
		properties.Set("Percentage Used", fmt.Sprintf("%d%%", smartCtlData.NvmeSmartHealthInformationLog.PercentageUsed))
	}
	if smartCtlData.NvmeSmartHealthInformationLog.Temperature != 0 {
		properties.Set("Temperature", fmt.Sprintf("%d °C", smartCtlData.NvmeSmartHealthInformationLog.Temperature))
	}
	if smartCtlData.NvmeSmartHealthInformationLog.MediaErrors > 0 {
		properties.Set("Media Errors", fmt.Sprintf("%d", smartCtlData.NvmeSmartHealthInformationLog.MediaErrors))
	}

	util.PrintFormattedTableOrdered(fmt.Sprintf("SmartCtl"), properties)

}

func init() {
	Command.AddCommand(diskListCmd)
}
