package maps

import "golang.org/x/exp/constraints"

func GetValueIfExists[MK constraints.Ordered, MV, T any](m map[MK]MV, key MK) T {
	if v, ok := m[key]; ok {
		if vs, ok := any(v).(T); ok {
			return vs
		}
	}
	var v T
	return v
}

// GetString returns the string value corresponding to the key.
// It returns a blank string if the value doesn't exist or it is not a string.
func GetString[K constraints.Ordered, V any](m map[K]V, key K) string {
	return GetValueIfExists[K, V, string](m, key)
}

// GetStringSlice returns the string value corresponding to the key.
// It returns a blank string if the value doesn't exist or it is not a string.
func GetStringSlice[K constraints.Ordered, V any](m map[K]V, key K) []string {
	return GetValueIfExists[K, V, []string](m, key)
}

// GetInt64 returns the int64 value corresponding to the key.
// It returns 0 if the value doesn't exist or it is not a signed integer.
func GetInt64[K constraints.Ordered, V any](m map[K]V, key K) int64 {
	if v, ok := m[key]; ok {
		switch vv := any(v).(type) {
		case int:
			return int64(vv)
		case int8:
			return int64(vv)
		case int16:
			return int64(vv)
		case int32:
			return int64(vv)
		case int64:
			return vv
		}
	}
	return 0
}
