package internal

import "strings"

func IsOneOf(list []string, s string, caseSensitive bool) bool {
	op := func(s string) string { return s }
	if !caseSensitive {
		op = strings.ToLower
	}
	s = op(s)
	for _, e := range list {
		if s == op(e) {
			return true
		}
	}
	return false
}
