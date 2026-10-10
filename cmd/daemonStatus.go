package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const defaultStatusWidth = 100

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the state of the running daemon",
	Long: `Shows what the daemon is watching and how often commands ran, how long they took and
how many were dropped or failed.

With --watch the view is interactive and updates continuously, press q to leave it.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		watch, _ := cmd.Flags().GetBool("watch")
		asJSON, _ := cmd.Flags().GetBool("json")
		interval, _ := cmd.Flags().GetDuration("interval")
		colorMode, _ := cmd.Flags().GetString("color")

		color, err := useColor(colorMode)
		if err != nil {
			return err
		}

		if watch {
			if asJSON {
				return fmt.Errorf("--json can not be combined with --watch")
			}
			if !isTerminal(os.Stdout) || !isTerminal(os.Stdin) {
				return fmt.Errorf("--watch needs a terminal")
			}
			return runLiveStatus(max(interval, 100*time.Millisecond), color)
		}

		info, legacy, err := daemon.FetchStatus()
		if err != nil {
			return err
		}
		switch {
		case info == nil:
			// a daemon that predates the structured status
			for _, line := range legacy {
				fmt.Println(line)
			}
		case asJSON:
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(info)
		default:
			fmt.Print(daemon.RenderStatus(*info, daemon.RenderOptions{Color: color, Width: terminalWidth()}))
		}
		return nil
	},
}

func init() {
	daemonStatusCmd.Flags().BoolP("watch", "w", false, "Interactive view that updates continuously")
	daemonStatusCmd.Flags().Duration("interval", time.Second, "Update interval of the --watch view")
	daemonStatusCmd.Flags().Bool("json", false, "Print the status as JSON")
	daemonStatusCmd.Flags().String("color", "auto", "Use colors: auto, always or never")
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func useColor(mode string) (bool, error) {
	switch mode {
	case "always":
		return true, nil
	case "never":
		return false, nil
	case "auto":
		return isTerminal(os.Stdout) && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb", nil
	}
	return false, fmt.Errorf("invalid value for --color: %q (expected auto, always or never)", mode)
}

// terminalSize returns columns and rows of the terminal on stdout, or zeros.
func terminalSize() (cols, rows int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		return 0, 0
	}
	return int(ws.Col), int(ws.Row)
}

func terminalWidth() int {
	if cols, _ := terminalSize(); cols > 0 {
		return cols
	}
	return defaultStatusWidth
}

// runLiveStatus draws the status on the alternate screen until q is pressed.
func runLiveStatus(interval time.Duration, color bool) error {
	stdin := int(os.Stdin.Fd())
	if saved, err := unix.IoctlGetTermios(stdin, unix.TCGETS); err == nil {
		raw := *saved
		// keep ISIG, ctrl-c still interrupts
		raw.Lflag &^= unix.ECHO | unix.ICANON
		raw.Cc[unix.VMIN], raw.Cc[unix.VTIME] = 1, 0
		if err := unix.IoctlSetTermios(stdin, unix.TCSETS, &raw); err == nil {
			defer func() { _ = unix.IoctlSetTermios(stdin, unix.TCSETS, saved) }()
		}
	}

	fmt.Print("\x1b[?1049h\x1b[?25l")
	defer fmt.Print("\x1b[?25h\x1b[?1049l")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigs)

	keys := make(chan byte, 8)
	go func() {
		buf := make([]byte, 1)
		for {
			if n, err := os.Stdin.Read(buf); err != nil || n == 0 {
				return
			}
			keys <- buf[0]
		}
	}()

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	draw(interval, color)
	for {
		select {
		case <-sigs:
			return nil
		case k := <-keys:
			switch k {
			case 'q', 'Q':
				return nil
			case 'r', 'R':
				draw(interval, color)
			}
		case <-winch:
			draw(interval, color)
		case <-ticker.C:
			draw(interval, color)
		}
	}
}

func draw(interval time.Duration, color bool) {
	cols, rows := terminalSize()
	if cols == 0 {
		cols = defaultStatusWidth
	}

	var body string
	info, legacy, err := daemon.FetchStatus()
	switch {
	case err != nil:
		body = "\x1b[31m● daemon is not running\x1b[0m\n\x1b[2m" + err.Error() + "\x1b[0m\n"
		if !color {
			body = "● daemon is not running\n" + err.Error() + "\n"
		}
	case info == nil:
		body = strings.Join(legacy, "\n") + "\n"
	default:
		body = daemon.RenderStatus(*info, daemon.RenderOptions{Color: color, Width: cols})
	}

	footer := fmt.Sprintf("q quit · r refresh · every %s · %s", interval, time.Now().Format("15:04:05"))
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if rows > 2 && len(lines) > rows-2 {
		lines = append(lines[:rows-3], "…")
	}
	lines = append(lines, "", footer)

	var sb strings.Builder
	sb.WriteString("\x1b[H")
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteString("\x1b[K\r\n")
	}
	sb.WriteString("\x1b[J")
	fmt.Print(sb.String())
}
