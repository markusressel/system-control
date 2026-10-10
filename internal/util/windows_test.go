package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWindowList(t *testing.T) {
	output := `0x03a00003 -1 host Desktop
0x04200004  0 host Terminal — fish
0x04200010  2 N/A  Untitled Document 1`

	windows := ParseWindowList(output)

	assert.Equal(t, []Window{
		{ID: "0x03a00003", Desktop: "-1", Title: "Desktop"},
		{ID: "0x04200004", Desktop: "0", Title: "Terminal — fish"},
		{ID: "0x04200010", Desktop: "2", Title: "Untitled Document 1"},
	}, windows)
}

func TestParseWindowListIgnoresGarbage(t *testing.T) {
	assert.Empty(t, ParseWindowList("Cannot open display\n"))
}
