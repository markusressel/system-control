package redshift

import (
	"fmt"

	"github.com/spf13/cobra"
)

var gammaCmd = &cobra.Command{
	Use:   "gamma",
	Short: "Show current display gamma",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			lastSetGamma := getLastSetGamma(display)
			fmt.Println(lastSetGamma)
		}
		return nil
	},
}

func init() {
	Command.AddCommand(gammaCmd)
}
