package it

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func WithMap(tcx TestContext, fn func(m *hz.Map)) {
	name := NewUniqueObjectName("map")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetMap(ctx, name))
	fn(m)
}

func MapTester(t *testing.T, fn func(tcx TestContext, m *hz.Map)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithMap(tcx, func(m *hz.Map) {
			fn(tcx, m)
		})
	})
}
