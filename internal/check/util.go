package check

import "strings"

func ContainsString(list []string, s string) bool {
	s = strings.ToLower(s)
	for _, e := range list {
		if s == strings.ToLower(e) {
			return true
		}
	}
	return false
}
