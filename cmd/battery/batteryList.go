package battery

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"

	"github.com/elliotchance/orderedmap/v2"
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
		batteries, err := util.GetBatteryList()
		if err != nil {
			return err
		}

		filter := util.DeviceFilter{
			Path:         filterPath,
			Type:         filterType,
			Manufacturer: filterManufacturer,
			Model:        filterModel,
			Serial:       filterSerial,
		}

		var filtered []util.BatteryInfo
		for _, battery := range batteries {
			bType, _ := battery.GetType()
			if filter.Matches(battery.Path, bType, battery.Manufacturer, battery.Model, battery.SerialNumber) {
				filtered = append(filtered, battery)
			}
		}

		// sort entries
		slices.SortFunc(filtered, func(a, b util.BatteryInfo) int {
			return cmp.Or(
				// sort by battery name
				util.CompareIgnoreCase(a.Name, b.Name),
			)
		})

		for i, battery := range filtered {
			printBatteryInfo(battery)

			if i < len(filtered)-1 {
				fmt.Println()
			}
		}
		return nil
	},
}

func printBatteryInfo(battery util.BatteryInfo) {
	properties := orderedmap.NewOrderedMap[string, string]()

	bPresent, e := battery.IsPresent()
	bPresentText := ""
	if e == nil {
		bPresentText = strconv.FormatBool(bPresent)
	}
	properties.Set("Present", bPresentText)

	properties.Set("Path", battery.Path)
	bType, _ := battery.GetType()
	properties.Set("Type", bType)
	properties.Set("Manufacturer", battery.Manufacturer)
	properties.Set("Model", battery.Model)
	properties.Set("Serial", battery.SerialNumber)

	bCapacity, e := battery.GetCapacity()
	bCapacityText := ""
	if e == nil {
		bCapacityText = strconv.Itoa(int(bCapacity))
		bCapacityText = fmt.Sprintf("%v %%", bCapacityText)
	}
	properties.Set("Capacity", bCapacityText)

	bCapacityLevel, _ := battery.GetCapacityLevel()
	properties.Set("Capacity Level", bCapacityLevel)

	bCycleCount, e := battery.GetCycleCount()
	bCycleCountText := ""
	if e == nil {
		bCycleCountText = strconv.Itoa(int(bCycleCount))
	}
	properties.Set("Cycle Count", bCycleCountText)

	bEnergyFullDesign, e := battery.GetEnergyFullDesign()
	bEnergyFullDesignText := ""
	if e == nil {
		bEnergyFullDesignText = fmt.Sprintf("%v Wh", util.RoundToTwoDecimals(bEnergyFullDesign))
	}
	properties.Set("Energy Full Design", bEnergyFullDesignText)

	bEnergyFull, e := battery.GetEnergyFull()
	bEnergyFullText := ""
	if e == nil {
		bEnergyFullText = fmt.Sprintf("%v Wh", util.RoundToTwoDecimals(bEnergyFull))
	}
	properties.Set("Energy Full", bEnergyFullText)

	bEnergyNow, e := battery.GetEnergyNow()
	bEnergyNowText := ""
	if e == nil {
		bEnergyNowText = fmt.Sprintf("%v Wh", util.RoundToTwoDecimals(bEnergyNow))
	}
	properties.Set("Energy Now", bEnergyNowText)

	degradation, e := battery.GetDegradation()
	degradationText := ""
	if e == nil {
		degradationText = fmt.Sprintf("%v %%", util.RoundToTwoDecimals(degradation))
	}
	properties.Set("Degradation", degradationText)

	bPowerNow, e := battery.GetPowerNow()
	bPowerNowText := ""
	if e == nil {
		bPowerNowText = fmt.Sprintf("%v W", util.RoundToTwoDecimals(bPowerNow))
	}
	properties.Set("Power Now", bPowerNowText)

	bOnline, e := battery.IsOnline()
	bOnlineText := ""
	if e == nil {
		bOnlineText = strconv.FormatBool(bOnline)
	}
	properties.Set("Online", bOnlineText)

	bStatus, _ := battery.GetStatus()
	properties.Set("Status", bStatus)
	properties.Set("Scope", battery.Scope)
	bTechnology, _ := battery.GetTechnology()
	properties.Set("Technology", bTechnology)

	util.PrintFormattedTableOrdered(battery.Name, properties)
}

func init() {
	batteryListCmd.Flags().StringVar(&filterPath, "path", "", "Filter by path (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterType, "type", "", "Filter by type (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterManufacturer, "manufacturer", "", "Filter by manufacturer (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterModel, "model", "", "Filter by model (supports globbing)")
	batteryListCmd.Flags().StringVar(&filterSerial, "serial", "", "Filter by serial number (supports globbing)")
	Command.AddCommand(batteryListCmd)
}
