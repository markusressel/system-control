package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/markusressel/system-control/cmd/bluetooth"
	"github.com/markusressel/system-control/internal/audio"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"

	// registers the display ops
	_ "github.com/markusressel/system-control/cmd/display/redshift"
	_ "github.com/markusressel/system-control/internal/backlight"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the system-control daemon",
	Long: `Starts a daemon that runs commands without the process startup and graph queries
of a regular invocation, and serves --watch subscriptions.

The daemon is optional: with no daemon running (or for commands it does not handle),
every command works on its own.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// flags and arguments of ops are taken from the commands that implement them
		if err := daemon.Default.Bind(RootCmd); err != nil {
			return err
		}

		eng := daemon.NewEngine(daemon.Default)
		eng.AddSource(audio.EventSource)
		eng.AddSource(bluetooth.EventSource)
		eng.AddStatus(audio.Status)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		if err := eng.Start(ctx); err != nil {
			return err
		}
		fmt.Println("Daemon started on", daemon.SocketPath())

		// SIGHUP (systemctl --user reload system-control) reads the configuration file again
		reload := make(chan os.Signal, 1)
		signal.Notify(reload, syscall.SIGHUP)
		defer signal.Stop(reload)

		for {
			select {
			case <-reload:
				if err := configuration.Reload(); err != nil {
					fmt.Fprintln(os.Stderr, "Could not reload the configuration:", err)
					continue
				}
				fmt.Println("Configuration reloaded")
				eng.Refresh("")
			case <-ctx.Done():
				fmt.Println("Shutting down...")
				return nil
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(DaemonCmd)
	DaemonCmd.AddCommand(daemonStatusCmd)
}
