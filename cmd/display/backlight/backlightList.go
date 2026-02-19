package backlight

import (
	"fmt"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var backlightListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current display brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		backlights, err := util.GetBacklights()
		if err != nil {
			return err
		}

		for i, backlight := range backlights {
			brightness, _ := backlight.GetBrightness()
			maxBrightness, _ := backlight.GetMaxBrightness()

			properties := orderedmap.NewOrderedMap[string, string]()
			properties.Set("Brightness", fmt.Sprintf("%d", brightness))
			properties.Set("MaxBrightness", fmt.Sprintf("%d", maxBrightness))

			util.PrintFormattedTableOrdered(backlight.Name, properties)

			if i < len(backlights)-1 {
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(backlightListCmd)
}
