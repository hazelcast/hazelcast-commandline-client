package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

func TestTypes(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "setAdd", f: setAddTest},
		{name: "setEmpty", f: setEmptyTest},
		{name: "setNew", f: setNewTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func setEmptyTest(t *testing.T) {
	s := types.NewSet[string]()
	assert.Equal(t, 0, s.Len())
}

func setNewTest(t *testing.T) {
	s := types.NewSet[string]("foo", "bar")
	assert.Equal(t, 2, s.Len())
	assert.True(t, s.Has("foo"))
	assert.True(t, s.Has("bar"))
}

func setAddTest(t *testing.T) {
	s := types.NewSet[string]()
	s.Add("foo")
	assert.Equal(t, 1, s.Len())
	assert.True(t, s.Has("foo"))
}
