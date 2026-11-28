package persistence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummy struct {
	Text   string
	Number int
}

func useTempDir(t *testing.T) {
	previous := BaseDir
	BaseDir = filepath.Join(t.TempDir(), "persistence")
	t.Cleanup(func() { BaseDir = previous })
}

func TestStructRoundTrip(t *testing.T) {
	useTempDir(t)
	test := dummy{Text: "hello", Number: 3}

	assert.NoError(t, SaveStruct("key", test))

	var value dummy
	assert.NoError(t, ReadStruct("key", &value))
	assert.Equal(t, test, value)
}

func TestIntAndFloatRoundTrip(t *testing.T) {
	useTempDir(t)

	assert.NoError(t, SaveInt("i", 6500))
	i, err := ReadInt("i")
	assert.NoError(t, err)
	assert.Equal(t, int64(6500), i)

	assert.NoError(t, SaveFloat("f", 0.25))
	f, err := ReadFloat("f")
	assert.NoError(t, err)
	assert.Equal(t, 0.25, f)
}

func TestOverwriteLeavesNoTempFiles(t *testing.T) {
	useTempDir(t)

	for i := 0; i < 5; i++ {
		assert.NoError(t, SaveInt("a", i))
	}

	entries, err := os.ReadDir(BaseDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestReadMissingKey(t *testing.T) {
	useTempDir(t)
	_, err := ReadInt("missing")
	assert.Error(t, err)
}

func TestInvalidKeysAreRejected(t *testing.T) {
	useTempDir(t)
	for _, key := range []string{"", "..", "../x", "a/b"} {
		assert.Error(t, SaveInt(key, 1), key)
		_, err := ReadInt(key)
		assert.Error(t, err, key)
	}
}
