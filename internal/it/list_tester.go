package it

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func WithList(tcx TestContext, fn func(m *hz.List)) {
	name := NewUniqueObjectName("list")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetList(ctx, name))
	fn(m)
}

func ListTester(t *testing.T, fn func(tcx TestContext, m *hz.List)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithList(tcx, func(m *hz.List) {
			fn(tcx, m)
		})
	})
}
