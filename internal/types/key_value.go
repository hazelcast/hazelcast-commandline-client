package types

import "golang.org/x/exp/constraints"

// TODO: consolidate KeyValue with Pair2

type KeyValue[K constraints.Ordered, V any] struct {
	Key   K
	Value V
}

type KeyValues[K constraints.Ordered, V any] []KeyValue[K, V]

func (kvs KeyValues[K, V]) Map() map[K]V {
	m := make(map[K]V, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Value
	}
	return m
}
