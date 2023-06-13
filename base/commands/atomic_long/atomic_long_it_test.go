package atomiclong_test

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestAtomicLong(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Get_Noninteractive", f: get_NonInteractiveTest},
		{name: "Set_NonInteractive", f: set_NonInteractiveTest},
		{name: "IncrementGet_NonInteractive", f: incrementGet_NonInteractiveTest},
		{name: "DecrementGet_NonInteractive", f: decrementGet_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}
func get_NonInteractiveTest(t *testing.T) {
	it.AtomicLongTester(t, func(tcx it.TestContext, al *hz.AtomicLong) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "atomic-long", "-n", al.Name(), "get"))
			tcx.AssertStdoutEquals("0\n")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(al.Set(ctx, 100))
			check.Must(tcx.CLC().Execute(ctx, "atomic-long", "-n", al.Name(), "get"))
			tcx.AssertStdoutEquals("100\n")
		})
	})
}

func set_NonInteractiveTest(t *testing.T) {
	it.AtomicLongTester(t, func(tcx it.TestContext, al *hz.AtomicLong) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "atomic-long", "-n", al.Name(), "set", "100")
			v := check.MustValue(al.Get(ctx))
			require.Equal(t, int64(100), v)
		})
	})
}

func incrementGet_NonInteractiveTest(t *testing.T) {
	it.AtomicLongTester(t, func(tcx it.TestContext, al *hz.AtomicLong) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(al.Set(ctx, 100))
			tcx.CLCExecute(ctx, "atomic-long", "-n", al.Name(), "increment-get")
			v := check.MustValue(al.Get(ctx))
			require.Equal(t, int64(101), v)
		})
		tcx.WithReset(func() {
			check.Must(al.Set(ctx, 100))
			tcx.CLCExecute(ctx, "atomic-long", "-n", al.Name(), "increment-get", "--by", "50")
			v := check.MustValue(al.Get(ctx))
			require.Equal(t, int64(150), v)
		})
	})
}

func decrementGet_NonInteractiveTest(t *testing.T) {
	it.AtomicLongTester(t, func(tcx it.TestContext, al *hz.AtomicLong) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(al.Set(ctx, 100))
			tcx.CLCExecute(ctx, "atomic-long", "-n", al.Name(), "decrement-get")
			v := check.MustValue(al.Get(ctx))
			require.Equal(t, int64(99), v)
		})
		tcx.WithReset(func() {
			check.Must(al.Set(ctx, 100))
			tcx.CLCExecute(ctx, "atomic-long", "-n", al.Name(), "decrement-get", "--by", "50")
			v := check.MustValue(al.Get(ctx))
			require.Equal(t, int64(50), v)
		})
	})
}

func objectExists(sn, name string, objects []types.DistributedObjectInfo) bool {
	for _, obj := range objects {
		if sn == obj.ServiceName && name == obj.Name {
			return true
		}
	}
	return false
}
