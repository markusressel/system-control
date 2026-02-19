package keyboard

import (
	"fmt"
	"strconv"

	"github.com/markusressel/system-control/internal/util"

	"github.com/spf13/cobra"
)

var keyboardBrightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Get/Set the current keyboard backlight brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			p, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			_, err = util.SetKeyboardBrightness(p)
			if err != nil {
				return err
			}
		} else {
			brightness, err := util.GetKeyboardBrightness()
			if err != nil {
				return err
			}
			fmt.Println(brightness)
		}

		return nil
	},
}

func init() {
	Command.AddCommand(keyboardBrightnessCmd)
}
