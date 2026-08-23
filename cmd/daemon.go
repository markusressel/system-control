package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/markusressel/system-control/cmd/display/redshift"
	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the system-control daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := daemon.NewEngine()
		if err != nil {
			return err
		}

		eng.RegisterModule(audio.NewDaemonModule())
		eng.RegisterModule(redshift.NewDaemonModule())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := eng.Start(ctx); err != nil {
			return err
		}

		fmt.Println("Daemon started on", eng.GetSocketPath())

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		fmt.Println("Shutting down...")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(DaemonCmd)
}
