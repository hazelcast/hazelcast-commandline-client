//go:build std || demo

package demo_test

import (
	"context"
	"fmt"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/require"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestGenerateData(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "generateData_Wikipedia", f: generateData_WikipediaTest},
		{name: "generateData_Wikipedia_MaxValues", f: generateData_Wikipedia_MaxValues_Test},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func generateData_WikipediaTest(t *testing.T) {
	it.MarkFlaky(t, "https://github.com/hazelcast/hazelcast-commandline-client/issues/350")
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			err := tcx.CLCExecuteErr(ctx, "demo", "generate-data", "wikipedia-event-stream", "map="+m.Name(), "--timeout", "2s", "--yes")
			require.Error(t, err)
			size := check.MustValue(m.Size(context.Background()))
			require.Greater(t, size, 0)
		})
	})
}

func generateData_Wikipedia_MaxValues_Test(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		count := 10
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "demo", "generate-data", "wikipedia-event-stream", "map="+m.Name(), fmt.Sprintf("--max-values=%d", count), "--yes")
			size := check.MustValue(m.Size(context.Background()))
			require.Equal(t, count, size)
		})
	})
}

func TestMapSetMany(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		count := 10
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "demo", "map-setmany", "10", "--name", m.Name(), "--size", "1")
			require.Equal(t, count, check.MustValue(m.Size(context.Background())))
			require.Equal(t, "a", check.MustValue(m.Get(ctx, "k1")))
		})
	})
}
