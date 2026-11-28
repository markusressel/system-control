package fan

import (
	"github.com/markusressel/system-control/internal/commandutil"
	internalfan "github.com/markusressel/system-control/internal/fan"
	"github.com/spf13/cobra"
)

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Get the current fan mode",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE:  commandutil.OpRunE("fan.mode"),
}

func init() {
	modeCmd.Flags().String(internalfan.TypeFlag, internalfan.TypeText, "output format: text or number")
	commandutil.AddWatchFlag(modeCmd)
}
