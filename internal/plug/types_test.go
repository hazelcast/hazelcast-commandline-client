package plug

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProperties_Get(t *testing.T) {
	p := NewProperties()
	_, found := p.Get("non-existent")
	require.False(t, found)
	p.Set("existent", "foo")
	v, found := p.Get("existent")
	require.True(t, found)
	require.Equal(t, "foo", v)
	p.Push()
	v, found = p.Get("existent")
	require.True(t, found)
	require.Equal(t, "foo", v)
	p.Set("existent", "bar")
	v, found = p.Get("existent")
	require.True(t, found)
	require.Equal(t, "bar", v)
	p.Pop()
	v, found = p.Get("existent")
	require.True(t, found)
	require.Equal(t, "foo", v)
}
