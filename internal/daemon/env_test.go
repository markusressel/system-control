package daemon

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/markusressel/system-control/internal/util"
)

func TestEnsureEnvironment(t *testing.T) {
	op := &Op{Name: "test.x11", Requires: []string{"DISPLAY"}}

	fakeManager := func(out string) {
		restore := util.SetRunner(func(ctx context.Context, command string, args ...string) (string, error) {
			return out, nil
		})
		t.Cleanup(restore)
	}

	// as a daemon
	runningEngine.Store(&Engine{})
	t.Cleanup(func() { runningEngine.Store(nil) })

	t.Run("imports from the service manager", func(t *testing.T) {
		t.Setenv("DISPLAY", "")
		t.Setenv("XAUTHORITY", "")
		fakeManager("PATH=/usr/bin\nDISPLAY=:7\nXAUTHORITY=/run/xauth\nOTHER=x")
		if err := op.ensureEnvironment(); err != nil {
			t.Fatal(err)
		}
		if got := os.Getenv("DISPLAY"); got != ":7" {
			t.Errorf("DISPLAY = %q", got)
		}
		if got := os.Getenv("XAUTHORITY"); got != "/run/xauth" {
			t.Errorf("XAUTHORITY = %q", got)
		}
		if got := os.Getenv("OTHER"); got != "" {
			t.Errorf("unrelated variable imported: %q", got)
		}
	})

	t.Run("unhandled without DISPLAY anywhere", func(t *testing.T) {
		t.Setenv("DISPLAY", "")
		fakeManager("PATH=/usr/bin")
		if err := op.ensureEnvironment(); !errors.Is(err, ErrUnhandled) {
			t.Errorf("err = %v, want ErrUnhandled", err)
		}
	})

	t.Run("keeps an existing value", func(t *testing.T) {
		t.Setenv("DISPLAY", ":1")
		fakeManager("DISPLAY=:9")
		if err := op.ensureEnvironment(); err != nil || os.Getenv("DISPLAY") != ":1" {
			t.Errorf("err = %v, DISPLAY = %q", err, os.Getenv("DISPLAY"))
		}
	})

	t.Run("CLI is never touched", func(t *testing.T) {
		runningEngine.Store(nil)
		t.Setenv("DISPLAY", "")
		if err := op.ensureEnvironment(); err != nil {
			t.Errorf("err = %v", err)
		}
	})
}

func TestArgsTypedAccessors(t *testing.T) {
	a := Args{Flags: map[string]string{"n": "5", "f": "0.25", "d": "2s", "bad": "x"}, Pos: []string{"p"}}

	if v, err := a.Int("n"); err != nil || v != 5 {
		t.Errorf("Int = %v, %v", v, err)
	}
	if v, err := a.Float("f"); err != nil || v != 0.25 {
		t.Errorf("Float = %v, %v", v, err)
	}
	if v, err := a.Duration("d"); err != nil || v.String() != "2s" {
		t.Errorf("Duration = %v, %v", v, err)
	}
	if v, err := a.Arg(0); err != nil || v != "p" {
		t.Errorf("Arg(0) = %v, %v", v, err)
	}

	if _, err := a.Arg(1); err == nil {
		t.Error("Arg(1) must fail")
	}
	for name, get := range map[string]func() error{
		"bad int":      func() error { _, err := a.Int("bad"); return err },
		"bad float":    func() error { _, err := a.Float("bad"); return err },
		"bad duration": func() error { _, err := a.Duration("bad"); return err },
		"unknown flag": func() error { _, err := a.Int("nope"); return err },
	} {
		if get() == nil {
			t.Errorf("%s must fail", name)
		}
	}
}

func TestExecute_LockFor(t *testing.T) {
	t.Setenv(EnvRuntimeDir, t.TempDir())

	unlock, err := AcquireLock("busy")
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	ran := 0
	op := &Op{
		Name: "test.mixed",
		LockFor: func(a Args) string {
			if a.Flag("set") == "" {
				return ""
			}
			return "busy"
		},
		Run: func(ctx context.Context, a Args) ([]string, error) { ran++; return nil, nil },
	}

	// a query ignores the busy lock
	if _, err := op.Execute(context.Background(), Args{Flags: map[string]string{}}, time.Now()); err != nil || ran != 1 {
		t.Errorf("query: ran=%d err=%v", ran, err)
	}
	// a mutation is dropped, the lock stays busy
	if _, err := op.Execute(context.Background(), Args{Flags: map[string]string{"set": "1"}}, time.Now()); err != nil || ran != 1 {
		t.Errorf("mutation: ran=%d err=%v", ran, err)
	}
}

func TestStatus_ReportsExecutions(t *testing.T) {
	var counter atomic.Int64
	eng := startTestEngine(t, &counter)
	eng.AddStatus(func() []string { return []string{"extra: 1"} })

	var out bytes.Buffer
	if handled, code := TryRemote([]string{"test", "echo", "fail"}, &out, &out); !handled || code != 1 {
		t.Fatalf("handled=%v code=%d", handled, code)
	}

	joined := strings.Join(eng.status(), "\n")
	for _, want := range []string{"Executions", "test.echo", "last error", "boom", "extra: 1"} {
		if !strings.Contains(joined, want) {
			t.Errorf("status lacks %q:\n%s", want, joined)
		}
	}
}
