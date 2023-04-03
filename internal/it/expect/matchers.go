package expect

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func Exact(s string) ExactMatcher {
	return ExactMatcher{pattern: s}
}

func Contains(s string) ContainsMatcher {
	return ContainsMatcher{pattern: s}
}

func Dollar(s string) DollarMatcher {
	return DollarMatcher{pattern: s}
}

type Matcher interface {
	Match(s string) bool
}

type ExactMatcher struct {
	pattern string
}

func (m ExactMatcher) Match(s string) bool {
	return m.pattern == s
}

type ContainsMatcher struct {
	pattern string
}

func (c ContainsMatcher) Match(s string) bool {
	return strings.Contains(s, c.pattern)
}

type DollarMatcher struct {
	pattern string
}

func (m DollarMatcher) Match(s string) bool {
	// normalize the pattern
	p := m.normalize(m.pattern)
	return check.MustValue(regexp.Match(p, []byte(s)))
}

func (m DollarMatcher) normalize(s string) string {
	// 1. trim spaces before and after $, if $ appears at the beginng or the end
	// 2. replace $ with \b+
	var lines []string
	scn := bufio.NewScanner(strings.NewReader(s))
	for scn.Scan() {
		line := strings.TrimSpace(scn.Text())
		line = strings.ReplaceAll(line, "$", "\\s*")
		line = strings.ReplaceAll(line, "[", "\\[")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
