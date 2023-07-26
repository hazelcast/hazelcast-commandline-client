package it

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func WithMultiMap(tcx TestContext, fn func(m *hz.MultiMap)) {
	name := NewUniqueObjectName("multiMap")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetMultiMap(ctx, name))
	fn(m)
}

func MultiMapTester(t *testing.T, fn func(tcx TestContext, m *hz.MultiMap)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithMultiMap(tcx, func(m *hz.MultiMap) {
			fn(tcx, m)
		})
	})
}
