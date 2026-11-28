package battery

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"sync"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
)

// PrintList prints the details of every battery matching `filter`, sorted by name.
func PrintList(filter util.DeviceFilter) error {
	batteries, err := GetBatteryList()
	if err != nil {
		return err
	}

	var filtered []BatteryInfo
	for _, battery := range batteries {
		bType, _ := battery.GetType()
		if filter.Matches(battery.Path, bType, battery.Manufacturer, battery.Model, battery.SerialNumber) {
			filtered = append(filtered, battery)
		}
	}

	// sort entries
	slices.SortFunc(filtered, func(a, b BatteryInfo) int {
		return cmp.Or(
			// sort by battery name
			util.CompareIgnoreCase(a.Name, b.Name),
		)
	})

	// Query HID++ battery details in parallel
	var wg sync.WaitGroup
	for i := range filtered {
		wg.Add(1)
		go func(b *BatteryInfo) {
			defer wg.Done()
			b.ResolveHIDPP()
		}(&filtered[i])
	}
	wg.Wait()

	for i, battery := range filtered {
		printBatteryInfo(battery)

		if i < len(filtered)-1 {
			fmt.Println()
		}
	}
	return nil
}

func printBatteryInfo(battery BatteryInfo) {
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
	properties.Set("Cached", strconv.FormatBool(battery.IsCached()))
	properties.Set("Scope", battery.Scope)
	bTechnology, _ := battery.GetTechnology()
	properties.Set("Technology", bTechnology)

	util.PrintFormattedTableOrdered(battery.Name, properties)
}
