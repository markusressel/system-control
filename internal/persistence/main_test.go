package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummy struct {
	Text   string
	Number int
}

func TestSaveStruct(t *testing.T) {
	// GIVEN
	BaseDir = "./"
	key := "key"
	test := dummy{
		Text:   "hello",
		Number: 0,
	}

	// WHEN
	err := SaveStruct(key, test)

	// THEN
	assert.NoError(t, err)
}

func TestReadStruct(t *testing.T) {
	// GIVEN
	BaseDir = "./"
	key := "key"
	test := dummy{
		Text:   "hello",
		Number: 0,
	}
	_ = SaveStruct(key, test)

	var value = dummy{}

	// WHEN
	err := ReadStruct(key, &value)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, test, value)
}
