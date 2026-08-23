package commandutil

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// StateCommandRunE wraps a state fetching command to transparently use the daemon if available.
// isMutating can be provided for commands that act as both getters and setters. If true, IPC is bypassed.
func StateCommandRunE(commandId string, isMutating func(cmd *cobra.Command) bool, fallback func(cmd *cobra.Command) (string, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		watch, _ := cmd.Flags().GetBool("watch")

		flags := make(map[string]string)
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			if flag.Name != "watch" && flag.Name != "configuration" {
				flags[flag.Name] = flag.Value.String()
			}
		})

		// Bypass daemon if this invocation is mutating state
		if isMutating == nil || !isMutating(cmd) {

			u, err := user.Current()
			if err == nil {
				socketPath := filepath.Join("/run/user", u.Uid, "system-control", "ipc.sock")
				client := daemon.NewIpcClient(socketPath)

				if client.IsActive() {
					req := daemon.IpcRequest{
						Command: commandId,
						Flags:   flags,
					}

					filePath, err := client.Subscribe(req)
					if err == nil {
						if watch {
							return watchFile(filePath)
						} else {
							// Wait a tiny bit for daemon to populate it if it's new
							time.Sleep(10 * time.Millisecond)
							if b, err := os.ReadFile(filePath); err == nil {
								fmt.Println(strings.TrimSpace(string(b)))
								return nil
							}
						}
					} else {
						fmt.Fprintf(os.Stderr, "Daemon subscribe error: %v\n", err)
					}
				} else {
					fmt.Fprintf(os.Stderr, "Daemon IPC socket not found at %s\n", socketPath)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Daemon error: failed to get current user: %v\n", err)
			}
		}

		// Fallback
		res, err := fallback(cmd)
		if err != nil {
			return err
		}
		fmt.Println(res)

		if isMutating != nil && isMutating(cmd) {
			PushState(commandId, flags, res)
		}

		if watch {
			// If daemon is not active but they requested watch, we can fallback to polling
			// but for now let's just return an error so they know the daemon is required for watching.
			return fmt.Errorf("daemon is not running, continuous watching is disabled")
		}

		return nil
	}
}

func watchFile(filePath string) error {
	// First, print the current content (if exists)
	if b, err := os.ReadFile(filePath); err == nil {
		fmt.Println(strings.TrimSpace(string(b)))
	} else {
		// Wait for the file to be created
		for i := 0; i < 150; i++ {
			time.Sleep(20 * time.Millisecond)
			if b, err := os.ReadFile(filePath); err == nil {
				fmt.Println(strings.TrimSpace(string(b)))
				break
			}
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	dirPath := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	os.MkdirAll(dirPath, 0755)

	if err := watcher.Add(dirPath); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if filepath.Base(event.Name) == fileName {
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if b, err := os.ReadFile(filePath); err == nil {
						fmt.Println(strings.TrimSpace(string(b)))
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Println("error:", err)
		}
	}
}

// PushState optionally pushes an updated state value directly to the daemon if it is running.
func PushState(commandId string, flags map[string]string, value string) {
	u, err := user.Current()
	if err != nil {
		fmt.Println("PushState error:", err)
		return
	}
	socketPath := filepath.Join("/run/user", u.Uid, "system-control", "ipc.sock")
	client := daemon.NewIpcClient(socketPath)
	if client.IsActive() {
		req := daemon.IpcRequest{
			Command: commandId,
			Flags:   flags,
			Value:   value,
		}
		err = client.Push(req)
		if err != nil {
			fmt.Println("PushState error:", err)
		}
	} else {
		fmt.Println("PushState: IPC not active")
	}
}
