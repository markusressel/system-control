/*
 * system-control
 * Copyright (c) 2019. Markus Ressel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package cmd

import (
	"fmt"
	"github.com/markusressel/system-control/cmd/audio"
	"github.com/markusressel/system-control/cmd/battery"
	"github.com/markusressel/system-control/cmd/bluetooth"
	"github.com/markusressel/system-control/cmd/cpu"
	"github.com/markusressel/system-control/cmd/display"
	"github.com/markusressel/system-control/cmd/global"
	"github.com/markusressel/system-control/cmd/keyboard"
	"github.com/markusressel/system-control/cmd/mouse"
	"github.com/markusressel/system-control/cmd/network"
	"github.com/markusressel/system-control/cmd/network/wifi"
	"github.com/markusressel/system-control/cmd/touchpad"
	"github.com/markusressel/system-control/cmd/video"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/spf13/cobra"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "system-control",
	Short: "A utility to make common system actions a breeze.",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	cobra.OnInitialize(func() {
		configuration.InitConfig(global.CfgFile)
	})

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(audio.Command)
	RootCmd.AddCommand(battery.Command)
	RootCmd.AddCommand(bluetooth.Command)
	RootCmd.AddCommand(cpu.Command)
	RootCmd.AddCommand(display.Command)
	RootCmd.AddCommand(keyboard.Command)
	RootCmd.AddCommand(mouse.Command)
	RootCmd.AddCommand(touchpad.Command)
	RootCmd.AddCommand(video.Command)
	RootCmd.AddCommand(network.Command)
	RootCmd.AddCommand(wifi.Command)

	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "configuration", "", "configuration file (default is $HOME/.system-control.yaml)")
}

// initConfig reads in configuration file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use configuration file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search configuration in home directory with name ".system-control" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".system-control")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a configuration file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using configuration file:", viper.ConfigFileUsed())
	}
}
