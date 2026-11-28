package util

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecCommandReturnsStdoutOnFailure(t *testing.T) {
	output, err := runCommand(context.Background(), "sh", "-c", "echo partial; echo bad >&2; exit 3")

	assert.Equal(t, "partial", output)
	assert.ErrorContains(t, err, "bad")
}

func TestExecCommandTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := runCommand(ctx, "sleep", "5")

	assert.ErrorContains(t, err, "timed out")
}

func TestSetRunnerReplacesExecution(t *testing.T) {
	restore := SetRunner(func(ctx context.Context, command string, args ...string) (string, error) {
		return command + " " + args[0], nil
	})
	defer restore()

	output, err := ExecCommand("tool", "arg")

	assert.NoError(t, err)
	assert.Equal(t, "tool arg", output)
}
