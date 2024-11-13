package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecCommandAndFork executes a shell command with the given arguments
// and disowns the process, resulting in the process continuing to run
// even after the parent process (this golang application) has exited.
func ExecCommandAndFork(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

// ExecCommand executes a shell command with the given arguments
// and returns its stdout as a []byte.
// If an error occurs the content of stderr is printed
// and an error is returned.
func ExecCommand(command string, args ...string) (string, error) {
	//log.Printf("Executing command: %s %s", command, args)
	cmd := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(stderr.String())
		return "", err
	}

	result := stdout.String()
	result = strings.TrimSpace(result)

	return result, nil
}

func ExecCommandOneshot(timeout time.Duration, command string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := exec.CommandContext(ctx, command, args...).Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ctx.Err()
		}
		return err
	}
	return nil
}
