package pipewire

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
