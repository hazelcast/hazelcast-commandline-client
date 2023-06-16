package it

import (
	"context"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	hz "github.com/hazelcast/hazelcast-go-client"
)

func WithTopic(tcx TestContext, fn func(m *hz.Topic)) {
	name := NewUniqueObjectName("topic")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetTopic(ctx, name))
	fn(m)
}

func TopicTester(t *testing.T, fn func(tcx TestContext, m *hz.Topic)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithTopic(tcx, func(m *hz.Topic) {
			fn(tcx, m)
		})
	})
}
