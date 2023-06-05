package _multimap_test

import (
	"context"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/require"
)

func TestMultimap(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Put_NonInteractive", f: put_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func put_NonInteractiveTest(t *testing.T) {
	it.MultiMapTester(t, func(tcx it.TestContext, m *hazelcast.MultiMap) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "multimap", "-n", m.Name(), "put", "foo", "bar", "-q")
			tcx.CLCExecute(ctx, "multimap", "-n", m.Name(), "put", "foo", "bar2", "-q")
			tcx.AssertStderrEquals("")
			v := check.MustValue(m.Get(context.Background(), "foo"))
			require.Contains(t, v, "bar")
			require.Contains(t, v, "bar2")
		})
	})
}
