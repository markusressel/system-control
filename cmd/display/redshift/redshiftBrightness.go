package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Show current display brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			lastSetBrightness := getLastSetBrightness(display)
			fmt.Println(lastSetBrightness)
		}
		return nil
	},
}

func init() {
	Command.AddCommand(brightnessCmd)
}
