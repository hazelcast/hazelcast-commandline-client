package maps

import "golang.org/x/exp/constraints"

// GetString returns the string value corresponding to the key.
// It returns a blank string if the value doesn't exist or it is not a string.
func GetString[K constraints.Ordered, V any](m map[K]V, key K) string {
	if v, ok := m[key]; ok {
		if vs, ok := any(v).(string); ok {
			return vs
		}
	}
	return ""
}
