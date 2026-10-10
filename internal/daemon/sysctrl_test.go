package daemon

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// buildSysctrl compiles the C client, or skips the test if there is no compiler.
func buildSysctrl(t *testing.T) string {
	t.Helper()
	gcc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("no gcc, can't build the C client")
	}

	dir := t.TempDir()
	binary := filepath.Join(dir, "sysctrl")
	if out, err := exec.Command(gcc, "-O0", "-Wall", "-o", binary, "../../cmd/sysctrl/main.c").CombinedOutput(); err != nil {
		t.Fatalf("building sysctrl: %v\n%s", err, out)
	}

	// the client runs this binary for everything the daemon does not handle
	fallback := "#!/bin/sh\necho \"fallback NO_DAEMON=$" + EnvNoDaemon + " $*\"\n"
	if err := os.WriteFile(filepath.Join(dir, "system-control"), []byte(fallback), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return binary
}

func runSysctrl(t *testing.T, binary string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return out.String(), errOut.String(), exitCode
}

func TestSysctrl(t *testing.T) {
	binary := buildSysctrl(t)
	var counter atomic.Int64
	counter.Store(1)
	startTestEngine(t, &counter)

	t.Run("answer", func(t *testing.T) {
		stdout, _, code := runSysctrl(t, binary, "test", "echo", "-p", "P:", "x")
		if code != 0 || stdout != "P:x\na,b\n" {
			t.Errorf("code=%d stdout=%q", code, stdout)
		}
	})

	t.Run("error", func(t *testing.T) {
		stdout, stderr, code := runSysctrl(t, binary, "test", "echo", "fail")
		if code != 1 || stderr != "boom\n" || stdout != "" {
			t.Errorf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})

	t.Run("unhandled runs the binary", func(t *testing.T) {
		for _, args := range [][]string{{"test", "unhandled"}, {"test", "nope", "--flag"}} {
			stdout, _, code := runSysctrl(t, binary, args...)
			want := "fallback NO_DAEMON=1 " + strings.Join(args, " ") + "\n"
			if code != 0 || stdout != want {
				t.Errorf("%v: code=%d stdout=%q, want %q", args, code, stdout, want)
			}
		}
	})

	t.Run("watch", func(t *testing.T) {
		cmd := exec.Command(binary, "test", "counter", "--watch")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		lines := make(chan string, 8)
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				lines <- scanner.Text()
			}
			close(lines)
		}()

		expect := func(want string) {
			t.Helper()
			select {
			case got := <-lines:
				if got != want {
					t.Errorf("got %q, want %q", got, want)
				}
			case <-time.After(3 * time.Second):
				t.Fatalf("no line, want %q", want)
			}
		}

		expect("1")
		if _, _, code := runSysctrl(t, binary, "test", "bump"); code != 0 {
			t.Fatalf("bump failed with %d", code)
		}
		expect("2")
	})
}

func TestSysctrl_NoDaemonRunsTheBinary(t *testing.T) {
	binary := buildSysctrl(t)
	t.Setenv(EnvRuntimeDir, t.TempDir())

	stdout, _, code := runSysctrl(t, binary, "audio", "volume")

	if code != 0 || stdout != "fallback NO_DAEMON=1 audio volume\n" {
		t.Errorf("code=%d stdout=%q", code, stdout)
	}
}
