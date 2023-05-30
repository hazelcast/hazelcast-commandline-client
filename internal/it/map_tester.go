package it

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func WithMap(tcx TestContext, mn string, fn func(m *hz.Map)) {
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetMap(ctx, mn))
	fn(m)
}

func WithRandomMap(tcx TestContext, fn func(m *hz.Map)) {
	WithMap(tcx, NewUniqueObjectName("map"), fn)
}

func MapTesterWithName(t *testing.T, mn string, fn func(tcx TestContext, m *hz.Map)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithMap(tcx, mn, func(m *hz.Map) {
			fn(tcx, m)
		})
	})
}

func MapTester(t *testing.T, fn func(tcx TestContext, m *hz.Map)) {
	MapTesterWithName(t, NewUniqueObjectName("map"), fn)
}
