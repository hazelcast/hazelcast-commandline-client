package str_test

import (
	"fmt"
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

func TestSpacePaddedIntFormat(t *testing.T) {
	testCases := []struct {
		num int
		out string
	}{
		{num: 0, out: "%1d"},
		{num: 9, out: "%1d"},
		{num: 10, out: "%2d"},
		{num: 99, out: "%2d"},
		{num: 100, out: "%3d"},
		{num: 9999, out: "%4d"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("pad %d", tc.num), func(t *testing.T) {
			s := str.SpacePaddedIntFormat(tc.num)
			assert.Equal(t, tc.out, s)
		})
	}
}
