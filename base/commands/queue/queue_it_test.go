//go:build std || queue

package queue_test

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

func TestQueue(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Clear_NonInteractive", f: clear_NonInteractiveTest},
		{name: "Poll_Noninteractive", f: poll_NonInteractiveTest},
		{name: "Offer_NonInteractive", f: offer_NonInteractiveTest},
		{name: "Size_Noninteractive", f: size_NoninteractiveTest},
		{name: "Destroy_NonInteractiveTest", f: destroy_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func clear_NonInteractiveTest(t *testing.T) {
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.MustValue(q.Add(ctx, "foo"))
			require.Equal(t, 1, check.MustValue(q.Size(ctx)))
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "clear", "-q", "--yes"))
			require.Equal(t, 0, check.MustValue(q.Size(ctx)))
		})
	})
}

func poll_NonInteractiveTest(t *testing.T) {
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.MustValue(q.Add(ctx, "foo"))
			require.Equal(t, 1, check.MustValue(q.Size(ctx)))
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "poll", "--count", "1", "-q", "--show-type"))
			tcx.AssertStdoutEquals("foo\tSTRING\n")
			require.Equal(t, 0, check.MustValue(q.Size(ctx)))
		})
	})
}

func offer_NonInteractiveTest(t *testing.T) {
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "offer", "foo", "-q"))
			require.Equal(t, 1, check.MustValue(q.Size(ctx)))
		})
	})
}

func size_NoninteractiveTest(t *testing.T) {
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "size", "-q"))
			tcx.AssertStdoutEquals("0\n")
		})
		// set an entry
		tcx.WithReset(func() {
			check.MustValue(q.Add(ctx, "foo"))
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "size", "-q"))
			tcx.AssertStdoutEquals("1\n")
		})
	})
}

func destroy_NonInteractiveTest(t *testing.T) {
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "destroy", "--yes"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameQueue, q.Name(), objects))
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
