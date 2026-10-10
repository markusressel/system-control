package daemon

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// testRegistry returns a registry bound to a small command tree, like the real application.
func testRegistry(counter *atomic.Int64) *Registry {
	r := NewRegistry()

	r.Register(Op{
		Name: "test.echo",
		Run: func(ctx context.Context, a Args) ([]string, error) {
			if len(a.Pos) == 1 && a.Pos[0] == "fail" {
				return nil, errors.New("boom")
			}
			return []string{a.Flag("prefix") + strings.Join(a.Pos, ","), a.Flag("cols")}, nil
		},
	})

	r.Register(Op{
		Name:  "test.unhandled",
		Run:   func(ctx context.Context, a Args) ([]string, error) { return nil, fmt.Errorf("later: %w", ErrUnhandled) },
		Watch: func(Args) bool { return true },
	})

	r.Register(Op{
		Name: "test.counter",
		Run: func(ctx context.Context, a Args) ([]string, error) {
			return []string{fmt.Sprint(counter.Load())}, nil
		},
		Watch: func(Args) bool { return true },
	})

	r.Register(Op{
		Name:        "test.bump",
		Run:         func(ctx context.Context, a Args) ([]string, error) { counter.Add(1); return []string{"bumped"}, nil },
		Invalidates: []string{"test.counter"},
	})

	root := &cobra.Command{Use: "app"}
	root.PersistentFlags().String("configuration", "", "")
	group := &cobra.Command{Use: "test"}
	root.AddCommand(group)

	echo := &cobra.Command{Use: "echo", Args: cobra.MaximumNArgs(1), Run: func(*cobra.Command, []string) {}}
	echo.Flags().StringP("prefix", "p", "", "")
	echo.Flags().StringSliceP("cols", "c", []string{"a", "b"}, "")
	echo.AddCommand(&cobra.Command{Use: "child", Run: func(*cobra.Command, []string) {}})

	counterCmd := &cobra.Command{Use: "counter", Run: func(*cobra.Command, []string) {}}
	counterCmd.Flags().BoolP("watch", "w", false, "")

	unhandledCmd := &cobra.Command{Use: "unhandled", Run: func(*cobra.Command, []string) {}}
	unhandledCmd.Flags().BoolP("watch", "w", false, "")

	group.AddCommand(echo, counterCmd, unhandledCmd, &cobra.Command{Use: "bump", Run: func(*cobra.Command, []string) {}})

	if err := r.Bind(root); err != nil {
		panic(err)
	}
	return r
}

func TestMatchArgv(t *testing.T) {
	r := testRegistry(new(atomic.Int64))

	tests := []struct {
		name   string
		argv   []string
		ok     bool
		pos    []string
		flags  map[string]string
		watch  bool
		opName string
	}{
		{name: "plain", argv: []string{"test", "echo"}, ok: true, opName: "test.echo", flags: map[string]string{"prefix": "", "cols": "a,b"}},
		{name: "positional", argv: []string{"test", "echo", "x"}, ok: true, opName: "test.echo", pos: []string{"x"}},
		{name: "too many positionals", argv: []string{"test", "echo", "x", "y"}, ok: false},
		{name: "long flag value", argv: []string{"test", "echo", "--prefix", "P"}, ok: true, opName: "test.echo", flags: map[string]string{"prefix": "P"}},
		{name: "long flag equals", argv: []string{"test", "echo", "--prefix=P"}, ok: true, opName: "test.echo", flags: map[string]string{"prefix": "P"}},
		{name: "short flag value", argv: []string{"test", "echo", "-p", "P"}, ok: true, opName: "test.echo", flags: map[string]string{"prefix": "P"}},
		{name: "short flag attached", argv: []string{"test", "echo", "-pP"}, ok: true, opName: "test.echo", flags: map[string]string{"prefix": "P"}},
		{name: "list replaces default", argv: []string{"test", "echo", "-c", "x"}, ok: true, opName: "test.echo", flags: map[string]string{"cols": "x"}},
		{name: "list accumulates", argv: []string{"test", "echo", "-c", "x", "-c", "y"}, ok: true, opName: "test.echo", flags: map[string]string{"cols": "x,y"}},
		{name: "flag before positional", argv: []string{"test", "echo", "-p", "P", "x"}, ok: true, opName: "test.echo", pos: []string{"x"}, flags: map[string]string{"prefix": "P"}},
		{name: "double dash", argv: []string{"test", "echo", "--", "-x"}, ok: true, opName: "test.echo", pos: []string{"-x"}},
		{name: "unknown flag", argv: []string{"test", "echo", "--nope"}, ok: false},
		{name: "missing value", argv: []string{"test", "echo", "--prefix"}, ok: false},
		{name: "sub command of an op", argv: []string{"test", "echo", "child"}, ok: false},
		{name: "root flag", argv: []string{"test", "echo", "--configuration", "x"}, ok: false},
		{name: "unknown command", argv: []string{"test", "nope"}, ok: false},
		{name: "flag before command is left to cobra", argv: []string{"--prefix", "x", "test", "echo"}, ok: false},
		{name: "watch", argv: []string{"test", "counter", "--watch"}, ok: true, opName: "test.counter", watch: true},
		{name: "watch short", argv: []string{"test", "counter", "-w"}, ok: true, opName: "test.counter", watch: true},
		{name: "watch unsupported", argv: []string{"test", "echo", "--watch"}, ok: false},
		{name: "empty", argv: nil, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			op, a, ok := r.MatchArgv(tc.argv)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if !ok {
				return
			}
			if op.Name != tc.opName {
				t.Errorf("op = %s, want %s", op.Name, tc.opName)
			}
			if !slices.Equal(a.Pos, tc.pos) {
				t.Errorf("pos = %v, want %v", a.Pos, tc.pos)
			}
			if a.Watch != tc.watch {
				t.Errorf("watch = %v, want %v", a.Watch, tc.watch)
			}
			for k, v := range tc.flags {
				if a.Flags[k] != v {
					t.Errorf("flag %s = %q, want %q", k, a.Flags[k], v)
				}
			}
		})
	}
}

func startTestEngine(t *testing.T, counter *atomic.Int64) *Engine {
	t.Helper()
	t.Setenv(EnvRuntimeDir, t.TempDir())
	t.Setenv(EnvNoDaemon, "")

	eng := NewEngine(testRegistry(counter))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if err := eng.Start(ctx); err != nil {
		t.Fatal(err)
	}
	return eng
}

func TestTryRemote(t *testing.T) {
	var counter atomic.Int64
	startTestEngine(t, &counter)

	var stdout, stderr bytes.Buffer

	handled, code := TryRemote([]string{"test", "echo", "-p", "P:", "x"}, &stdout, &stderr)
	if !handled || code != 0 || stdout.String() != "P:x\na,b\n" {
		t.Errorf("handled=%v code=%d stdout=%q stderr=%q", handled, code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	handled, code = TryRemote([]string{"test", "echo", "fail"}, &stdout, &stderr)
	if !handled || code != 1 || stderr.String() != "boom\n" {
		t.Errorf("handled=%v code=%d stderr=%q", handled, code, stderr.String())
	}

	// an op that can't serve the request right now hands it back to the caller, also for watches
	for _, argv := range [][]string{{"test", "unhandled"}, {"test", "unhandled", "--watch"}} {
		stdout.Reset()
		stderr.Reset()
		if handled, _ := TryRemote(argv, &stdout, &stderr); handled || stdout.Len() > 0 || stderr.Len() > 0 {
			t.Errorf("%v: ErrUnhandled must be answered with \"unhandled\"", argv)
		}
	}

	// an error is an answer, only unknown commands go back to the caller
	handled, _ = TryRemote([]string{"test", "nope"}, &stdout, &stderr)
	if handled {
		t.Error("unknown command must not be handled")
	}
}

func TestTryRemote_NoDaemon(t *testing.T) {
	t.Setenv(EnvRuntimeDir, t.TempDir())
	t.Setenv(EnvNoDaemon, "")

	handled, _ := TryRemote([]string{"test", "echo"}, &bytes.Buffer{}, &bytes.Buffer{})
	if handled {
		t.Error("must not be handled without a daemon")
	}
}

func TestStart_RefusesSecondDaemon(t *testing.T) {
	var counter atomic.Int64
	startTestEngine(t, &counter)

	second := NewEngine(testRegistry(&counter))
	if err := second.Start(context.Background()); err == nil {
		t.Error("second daemon must not start")
	}
}

func readFrame(t *testing.T, r *bufio.Reader) string {
	t.Helper()
	got := make(chan string, 1)
	go func() {
		line, _ := r.ReadString('\n')
		got <- strings.TrimSuffix(line, "\n")
	}()
	select {
	case line := <-got:
		return line
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for frame")
		return ""
	}
}

func TestWatch(t *testing.T) {
	var counter atomic.Int64
	eng := startTestEngine(t, &counter)

	conn, err := net.Dial("unix", SocketPath())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "ARGV 3\n4\ntest7\ncounter7\n--watch")
	r := bufio.NewReader(conn)

	if got := readFrame(t, r); got != "W" {
		t.Fatalf("first frame = %q, want the watch announcement", got)
	}
	if got := readFrame(t, r); got != "O 0" {
		t.Fatalf("initial frame = %q", got)
	}

	// a state change followed by a refresh is pushed
	counter.Store(5)
	eng.Refresh("test.")
	if got := readFrame(t, r); got != "O 5" {
		t.Fatalf("update frame = %q", got)
	}

	// a mutating command refreshes what it invalidates
	var stdout bytes.Buffer
	if handled, _ := TryRemote([]string{"test", "bump"}, &stdout, &bytes.Buffer{}); !handled {
		t.Fatal("bump not handled")
	}
	if got := readFrame(t, r); got != "O 6" {
		t.Fatalf("frame after bump = %q", got)
	}

	// an unchanged value is not sent again
	eng.Refresh("test.")
	counter.Store(7)
	eng.Refresh("test.")
	if got := readFrame(t, r); got != "O 7" {
		t.Fatalf("frame = %q", got)
	}

	// the entry goes away with its last subscriber
	conn.Close()
	deadline := time.Now().Add(2 * time.Second)
	for {
		eng.store.mu.Lock()
		n := len(eng.store.entries)
		eng.store.mu.Unlock()
		if n == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d entries left after the subscriber disconnected", n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestExecute_DropsStaleQueuedCommands(t *testing.T) {
	t.Setenv(EnvRuntimeDir, t.TempDir())

	ran := false
	op := &Op{
		Name: "test.locked",
		Lock: "test-lock",
		Run:  func(ctx context.Context, a Args) ([]string, error) { ran = true; return []string{"x"}, nil },
	}

	lines, err := op.Execute(context.Background(), Args{}, time.Now().Add(-time.Second))
	if err != nil || ran || lines != nil {
		t.Errorf("stale command must be dropped: ran=%v lines=%v err=%v", ran, lines, err)
	}

	lines, err = op.Execute(context.Background(), Args{}, time.Now())
	if err != nil || !ran || len(lines) != 1 {
		t.Errorf("fresh command must run: ran=%v lines=%v err=%v", ran, lines, err)
	}
}
