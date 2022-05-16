package internal

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestNamePersistence(t *testing.T) {
	persister := NewNamePersister()
	// assert empty
	v, isSet := persister.Get("test")
	assert.Equal(t, v, "")
	assert.False(t, isSet)
	// assert value is set
	persister.Set("test", "value")
	v, isSet = persister.Get("test")
	assert.Equal(t, v, "value")
	assert.True(t, isSet)
	// assert value is updated
	persister.Set("test", "value2")
	v, isSet = persister.Get("test")
	assert.Equal(t, v, "value2")
	assert.True(t, isSet)
	// assert value is reset
	persister.Reset("test")
	v, isSet = persister.Get("test")
	assert.Equal(t, v, "")
	assert.False(t, isSet)
}
