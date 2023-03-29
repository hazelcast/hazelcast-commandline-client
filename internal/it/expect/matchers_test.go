package expect

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type testCase struct {
	name    string
	pattern string
	s       string
	matches bool
}

func TestExactMatcher_Match(t *testing.T) {
	testCases := []testCase{
		{pattern: "", s: "", matches: true},
		{pattern: "foo", s: "foo", matches: true},
		{pattern: "foo1", s: "foo", matches: false},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.pattern, func(t *testing.T) {
			m := Exact(tc.pattern)
			assert.Equal(t, tc.matches, m.Match(tc.s))
		})
	}
}

func TestContainsMatcher_Match(t *testing.T) {
	testCases := []testCase{
		{pattern: "", s: "", matches: true},
		{pattern: "foo", s: "foo", matches: true},
		{pattern: "foo", s: "abcdfooefg", matches: true},
		{pattern: "foo", s: "", matches: false},
		{pattern: "foo", s: "fo", matches: false},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.pattern, func(t *testing.T) {
			m := Contains(tc.pattern)
			assert.Equal(t, tc.matches, m.Match(tc.s))
		})
	}
}

func TestDollarMatcher_Match(t *testing.T) {
	pattern := loadPattern("map_size_0.txt")
	testCases := []testCase{
		{
			name:    "empty string",
			pattern: pattern,
			s:       strings.ReplaceAll(pattern, "$", ""),
			matches: true,
		},
		{
			name:    "single space",
			pattern: pattern,
			s:       strings.ReplaceAll(pattern, "$", " "),
			matches: true,
		},
		{
			name:    "single tab",
			pattern: pattern,
			s:       strings.ReplaceAll(pattern, "$", "\t"),
			matches: true,
		},
		{
			name:    "mix of spaces and tabs",
			pattern: pattern,
			s:       strings.ReplaceAll(pattern, "$", "\t     \t\t"),
			matches: true,
		},
		{
			name:    "non-space",
			pattern: pattern,
			s:       strings.ReplaceAll(pattern, "$", "%"),
			matches: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := Dollar(tc.pattern)
			assert.Equal(t, tc.matches, m.Match(tc.s))
		})
	}
}

func TestDollarMatcher_normalize(t *testing.T) {
	p1 := string(check.MustValue(os.ReadFile("testdata/map_size_0.txt")))
	p2 := DollarMatcher{}.normalize(p1)
	p1 = strings.ReplaceAll(p1, "$", "")
	assert.True(t, check.MustValue(regexp.Match(p2, []byte(p1))))
}

func loadPattern(name string) string {
	return string(check.MustValue(os.ReadFile(fmt.Sprintf("testdata/%s", name))))
}
