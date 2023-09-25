package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"

	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

func TestTypes(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "setAdd", f: setAddTest},
		{name: "setDiff", f: setDiffTest},
		{name: "setEmpty", f: setEmptyTest},
		{name: "setNew", f: setNewTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func setEmptyTest(t *testing.T) {
	s := types.MakeSet[string]()
	assert.Equal(t, 0, s.Len())
}

func setNewTest(t *testing.T) {
	s := types.MakeSet[string]("foo", "bar")
	assert.Equal(t, 2, s.Len())
	assert.True(t, s.Has("foo"))
	assert.True(t, s.Has("bar"))
}

func setAddTest(t *testing.T) {
	s := types.MakeSet[string]()
	s.Add("foo")
	assert.Equal(t, 1, s.Len())
	assert.True(t, s.Has("foo"))
}

func setDiffTest(t *testing.T) {
	s1 := types.MakeSet[string]("foo", "bar", "baz")
	s2 := types.MakeSet[string]("bar")
	s3 := s1.Diff(s2)
	items := s3.Items()
	slices.Sort(items)
	assert.Equal(t, []string{"baz", "foo"}, items)
}
