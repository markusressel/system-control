package display

import (
	"testing"

	"github.com/markusressel/system-control/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestCreateDisplayConfigurationsFromArgs(t *testing.T) {
	configs, err := createDisplayConfigurationsFromArgs(
		[]string{"DP-1", "DP-2"},
		[]string{"3840x2160"},
		[]string{"", "3840x0"},
		nil,
		[]bool{false, true},
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, []util.DisplayConfig{
		{Name: "DP-1", Mode: "3840x2160"},
		{Name: "DP-2", Position: "3840x0", Primary: true},
	}, configs)
}

func TestTooManyOptionValues(t *testing.T) {
	_, err := createDisplayConfigurationsFromArgs([]string{"DP-1"}, []string{"a", "b"}, nil, nil, nil, nil, nil)
	assert.Error(t, err)
}
