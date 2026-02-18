package redshift

import (
	"fmt"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var redshiftBrightnessDecCmd = &cobra.Command{
	Use:   "dec",
	Short: "Decrease the currently applied redshift brightness.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		//configPath := configuration.DetectAndReadConfigFile()
		////ui.Info("Using configuration file at: %s", configPath)
		//config := configuration.LoadConfig()
		//err = configuration.Validate(configPath)
		//if err != nil {
		//	//ui.FatalWithoutStacktrace(err.Error())
		//}
		//
		//redshiftConfig, err := util.ReadRedshiftConfig()
		//if err != nil {
		//	return err
		//}

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			lastSetBrightness := getLastSetBrightness(display)

			rawNew := lastSetBrightness - stepFloat
			rounded := util.RoundTo2DP(rawNew)
			newBrightness := util.Clamp(rounded, 0.1, 1.0)
			err = ApplyRedshift(display, -1, newBrightness, -1)
			if err != nil {
				return err
			}

			fmt.Println(newBrightness)
		}

		return nil
	},
}

func init() {
	redshiftBrightnessDecCmd.PersistentFlags().Float64VarP(
		&stepFloat,
		"step", "s",
		0.1,
		"Step size to increase the brightness by (between 0.1 and 1.0)",
	)

	brightnessCmd.AddCommand(redshiftBrightnessDecCmd)
}
