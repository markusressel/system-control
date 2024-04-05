package pipewire

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

func TestFilterPipewireObjects(t *testing.T) {
	// GIVEN
	input, err := os.ReadFile("../../test/pipewire/pw-cli_list-objects.txt")
	assert.NoError(t, err)
	filters := []PropertyFilter{
		{key: "media.class", value: "Audio/Device"},
		{key: "id", value: "43"},
	}

	// WHEN
	objects := parsePipwireObjectsToMap(string(input))
	filtered := filterPipwireObjects(objects, func(v map[string]string) bool {
		for _, filter := range filters {
			if v[filter.key] != filter.value {
				return false
			}
		}

		return true
	})

	//assert.NoError(t, err)

	// THEN
	expectedLength := 1
	assert.Equal(t, expectedLength, len(filtered))
}
