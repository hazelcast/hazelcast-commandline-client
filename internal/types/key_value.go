package types

import "golang.org/x/exp/constraints"

// TODO: consolidate KeyValue with Pair2

type KeyValue[K constraints.Ordered, V any] struct {
	Key   K
	Value V
}

type KeyValues[K constraints.Ordered, V any] []KeyValue[K, V]
