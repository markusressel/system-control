package battery

import (
	batteryutil "github.com/markusressel/system-control/internal/battery"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var (
	filterPath         string
	filterType         string
	filterManufacturer string
	filterModel        string
	filterSerial       string
)

var batteryListCmd = &cobra.Command{
	Use:   "list",
	Short: "Get a list of all known batteries",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return batteryutil.PrintList(util.DeviceFilter{
			Path:         filterPath,
			Type:         filterType,
			Manufacturer: filterManufacturer,
			Model:        filterModel,
			Serial:       filterSerial,
		})
	},
}

func init() {
	batteryListCmd.Flags().StringVar(&filterPath, "path", "", "Filter by path (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterType, "type", "", "Filter by type (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterManufacturer, "manufacturer", "", "Filter by manufacturer (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterModel, "model", "", "Filter by model (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterSerial, "serial", "", "Filter by serial number (supports globbing)")
	Command.AddCommand(batteryListCmd)
}
