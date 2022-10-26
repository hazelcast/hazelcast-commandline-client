package output_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

func TestOrderedSet_Items(t *testing.T) {
	s := output.NewOrderedSet[string]()
	require.Equal(t, []string{}, s.Items())
	require.False(t, s.Contains("foo"))
	require.Equal(t, 0, s.Len())
	require.True(t, s.Add("foo"))
	require.True(t, s.Contains("foo"))
	require.Equal(t, []string{"foo"}, s.Items())
	require.Equal(t, 1, s.Len())
	require.False(t, s.Add("foo"))
	require.Equal(t, 1, s.Len())
	require.True(t, s.Contains("foo"))
	require.Equal(t, []string{"foo"}, s.Items())
	require.True(t, s.Add("bar"))
	require.Equal(t, []string{"foo", "bar"}, s.Items())
	require.True(t, s.Contains("foo"))
	require.True(t, s.Contains("bar"))
	require.True(t, s.Delete("foo"))
	require.False(t, s.Contains("foo"))
	require.Equal(t, []string{"bar"}, s.Items())
	require.False(t, s.Delete("foo"))
	require.Equal(t, []string{"bar"}, s.Items())
	require.True(t, s.Add("baz"))
	require.Equal(t, []string{"bar", "baz"}, s.Items())
	require.True(t, s.Delete("baz"))
	require.Equal(t, []string{"bar"}, s.Items())
	require.False(t, s.Delete("baz"))
	require.Equal(t, []string{"bar"}, s.Items())
	require.True(t, s.Add("baz"))
	require.Equal(t, []string{"bar", "baz"}, s.Items())
	require.False(t, s.Add("baz"))
	require.Equal(t, []string{"bar", "baz"}, s.Items())
	require.False(t, s.Add("baz"))
	require.Equal(t, []string{"bar", "baz"}, s.Items())
	require.True(t, s.Delete("baz"))
	require.Equal(t, []string{"bar"}, s.Items())
}
