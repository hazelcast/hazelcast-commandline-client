package it

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func WithAtomicLong(tcx TestContext, fn func(al *hz.AtomicLong)) {
	name := NewUniqueObjectName("AtomicLong")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.CPSubsystem().GetAtomicLong(ctx, name))
	fn(m)
}

func AtomicLongTester(t *testing.T, fn func(tcx TestContext, al *hz.AtomicLong)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithAtomicLong(tcx, func(m *hz.AtomicLong) {
			fn(tcx, m)
		})
	})
}
