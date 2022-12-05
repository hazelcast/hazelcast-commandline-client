package clc

import "golang.org/x/exp/constraints"

type KeyValue[K, V any] struct {
	Key   K
	Value V
}

type KeyValues[K constraints.Ordered, V any] []KeyValue[K, V]
