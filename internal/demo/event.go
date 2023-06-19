package demo

import (
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

type StreamItem interface {
	ID() string
	Row() output.Row
	FlatMap() map[string]any
}
