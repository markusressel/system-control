package cpu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCpuInfo(t *testing.T) {
	content := `processor	: 0
vendor_id	: GenuineIntel
model name	: 11th Gen Intel(R) Core(TM) i9-11900H @ 2.50GHz
flags		: fpu vme

processor	: 1
vendor_id	: GenuineIntel
model name	: 11th Gen Intel(R) Core(TM) i9-11900H @ 2.50GHz

`

	cpus := ParseCpuInfo(content)

	assert.Equal(t, []CpuInfo{
		{Index: 0, VendorId: "GenuineIntel", ModelName: "11th Gen Intel(R) Core(TM) i9-11900H @ 2.50GHz"},
		{Index: 1, VendorId: "GenuineIntel", ModelName: "11th Gen Intel(R) Core(TM) i9-11900H @ 2.50GHz"},
	}, cpus)
}

func TestParseCpuInfoGarbage(t *testing.T) {
	assert.Empty(t, ParseCpuInfo(""))
	assert.Empty(t, ParseCpuInfo("no colon here"))
}
