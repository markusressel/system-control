package media

import (
	"context"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"
)

func fakeFollow(t *testing.T, script string) {
	t.Helper()
	previous, delay := followCommand, followRestartDelay
	followCommand = func(ctx context.Context) *exec.Cmd { return exec.CommandContext(ctx, "sh", "-c", script) }
	followRestartDelay = 10 * time.Millisecond
	t.Cleanup(func() { followCommand, followRestartDelay = previous, delay })
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestFollow_CallsBackForEveryLine(t *testing.T) {
	fakeFollow(t, "echo Playing; echo Paused; sleep 30")

	ctx, cancel := context.WithCancel(context.Background())
	var changes atomic.Int64
	done := make(chan struct{})
	go func() { Follow(ctx, func() { changes.Add(1) }); close(done) }()

	waitFor(t, "two changes", func() bool { return changes.Load() >= 2 })
	if !Following() {
		t.Error("Following() = false while the command runs")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Follow did not return after the context was canceled")
	}
	if Following() {
		t.Error("Following() = true after Follow returned")
	}
}

func TestFollow_RestartsACommandThatEnded(t *testing.T) {
	fakeFollow(t, "echo changed")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var changes atomic.Int64
	go Follow(ctx, func() { changes.Add(1) })

	// every run prints one line and ends
	waitFor(t, "three runs", func() bool { return changes.Load() >= 3 })
}

func TestFollow_SurvivesAMissingTool(t *testing.T) {
	previous, delay := followCommand, followRestartDelay
	followCommand = func(ctx context.Context) *exec.Cmd { return exec.CommandContext(ctx, "/nonexistent/playerctl") }
	followRestartDelay = 10 * time.Millisecond
	defer func() { followCommand, followRestartDelay = previous, delay }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() { Follow(ctx, func() {}); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Follow did not return")
	}
	if Following() {
		t.Error("Following() = true without a tool")
	}
}
