package demo_test

import (
	"context"
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func generateData_WikipediaTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			err := tcx.CLCExecuteErr(ctx, "demo", "generate-data", "wikipedia-event-stream", "map-name="+m.Name(), "--preview", "--timeout", "2s")
			require.Error(t, err)
			size := check.MustValue(m.Size(context.Background()))
			require.Greater(t, size, 0)
		})
	})
}
