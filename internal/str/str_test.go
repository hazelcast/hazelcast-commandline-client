package str_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
)

func TestSplitByComma(t *testing.T) {
	testCases := []struct {
		s           string
		removeEmpty bool
		target      []string
	}{
		{
			s:      "",
			target: nil,
		},
		{
			s:           "",
			removeEmpty: true,
			target:      nil,
		},
		{
			s:      "foo",
			target: []string{"foo"},
		},
		{
			s:           "foo",
			removeEmpty: true,
			target:      []string{"foo"},
		},
		{
			s:      "foo,bar",
			target: []string{"foo", "bar"},
		},
		{
			s:           "foo,bar",
			removeEmpty: true,
			target:      []string{"foo", "bar"},
		},
		{
			s:      "foo,  bar",
			target: []string{"foo", "bar"},
		},
		{
			s:           "foo,  bar",
			removeEmpty: true,
			target:      []string{"foo", "bar"},
		},
		{
			s:      "foo,  , bar",
			target: []string{"foo", "", "bar"},
		},
		{
			s:           "foo,  , bar",
			removeEmpty: true,
			target:      []string{"foo", "bar"},
		},
		{
			s:      ", bar",
			target: []string{"", "bar"},
		},
		{
			s:           ", bar",
			removeEmpty: true,
			target:      []string{"bar"},
		},
		{
			s:      "bar  ,  ",
			target: []string{"bar", ""},
		},
		{
			s:           "bar  ,  ",
			removeEmpty: true,
			target:      []string{"bar"},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			o := str.SplitByComma(tc.s, tc.removeEmpty)
			assert.Equal(t, tc.target, o)
		})
	}
}
