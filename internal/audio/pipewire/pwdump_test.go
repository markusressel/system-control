package pipewire

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePwDumpToPipewireState(t *testing.T) {
	// GIVEN
	input, err := os.ReadFile("../../../test/pipewire/pw.dump")
	assert.NoError(t, err)

	// WHEN
	result, err := parsePwDumpToState(string(input))
	assert.NoError(t, err)

	// THEN
	assert.Equal(t, result, result)
}
