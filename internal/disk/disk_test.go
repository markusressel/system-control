package disk

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func exitError(t *testing.T, code string) error {
	err := exec.Command("sh", "-c", "exit "+code).Run()
	assert.Error(t, err)
	return err
}

func TestSmartctlResultIsUsable(t *testing.T) {
	// bit 6: error log contains entries, but the output is valid
	assert.True(t, smartctlResultIsUsable(exitError(t, "64"), `{"smartctl":{}}`))
	// bit 1: device open failed
	assert.False(t, smartctlResultIsUsable(exitError(t, "2"), `{"smartctl":{}}`))
	// no json
	assert.False(t, smartctlResultIsUsable(exitError(t, "64"), "Smartctl open device failed"))
	// not an exit error at all
	assert.False(t, smartctlResultIsUsable(&exec.Error{Name: "sudo", Err: exec.ErrNotFound}, `{}`))
}
