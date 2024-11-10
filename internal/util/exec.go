package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

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

const DefaultColumnHeaderRegexPattern = "\\S+\\s*"

func ParseTable[T any](input string, cellSeparator string, producer func(row []string) T) ([]T, error) {
	result := make([]T, 0)

	lines := strings.Split(input, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid table format")
	}

	headerCellRegex, err := regexp.Compile(cellSeparator)
	if err != nil {
		return nil, err
	}
	header := headerCellRegex.FindAllString(lines[0], -1)
	if len(header) < 2 {
		return nil, fmt.Errorf("invalid table format")
	}

	for i := 1; i < len(lines); i++ {
		row := make([]string, 0)
		currentLine := lines[i]
		startIdx := 0
		for i := 0; i < len(header); i++ {
			endIdx := startIdx + len(header[i])
			columnValue := SubstringRunes(currentLine, startIdx, endIdx)
			row = append(row, columnValue)
			startIdx = endIdx
		}
		if len(row) < 2 {
			return nil, fmt.Errorf("invalid table format")
		}
		result = append(result, producer(row))
	}

	return result, nil
}
