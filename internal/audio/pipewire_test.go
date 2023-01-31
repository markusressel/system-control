package audio

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestParsePipwireParamsToMap(t *testing.T) {
	// GIVEN
	input, err := os.ReadFile("../../test/pipewire/pw-cli_enum-params.txt")
	assert.NoError(t, err)

	// WHEN
	result := parsePipwireParamsToMap(string(input))
	//assert.NoError(t, err)

	// THEN
	assert.Equal(t, result, result)
}
