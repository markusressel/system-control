/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"strings"
	"time"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Reboot the system gracefully",
	Long:  `Reboots the system gracefully by first closing all currently open windows.`,
	Run: func(cmd *cobra.Command, args []string) {
		openWindows := findOpenWindows()

		for _, element := range openWindows {
			windowId := strings.Split(element, " ")[0]
			_, err := execCommand("wmctrl", "-i", "-c", windowId)
			if err != nil {
				log.Fatal(err)
			}
		}

		// wait for all windows to disappear
		for {
			openWindows = findOpenWindows()
			if len(openWindows) <= 0 {
				break
			} else {
				time.Sleep(time.Second)
			}
		}

		_, err := execCommand("reboot")
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
