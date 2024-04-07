package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
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
