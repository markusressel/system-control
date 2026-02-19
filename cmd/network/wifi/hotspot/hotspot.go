package wifi

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "hotspot",
	Short: "Control WiFi hotspots",
	Long:  ``,
}

func init() {
	Command.PersistentFlags().StringVarP(
		&wifiInterface,
		"interface", "i",
		"",
		"WiFi interface to use for the hotspot",
	)
}
