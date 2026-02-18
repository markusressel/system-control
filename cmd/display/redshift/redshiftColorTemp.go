package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var colorTempCmd = &cobra.Command{
	Use:   "color-temperature",
	Short: "Show current display color temperature",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			lastSetColorTemperature := getLastSetColorTemperature(display)
			fmt.Println(lastSetColorTemperature)
		}
		return nil
	},
}

func init() {
	Command.AddCommand(colorTempCmd)
}
