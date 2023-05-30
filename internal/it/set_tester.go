package it

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func WithSet(tcx TestContext, fn func(m *hz.Set)) {
	name := NewUniqueObjectName("set")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetSet(ctx, name))
	fn(m)
}

func SetTester(t *testing.T, fn func(tcx TestContext, s *hz.Set)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithSet(tcx, func(s *hz.Set) {
			fn(tcx, s)
		})
	})
}
