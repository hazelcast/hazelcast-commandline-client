package _queue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Clear_NonInteractive", f: clear_NonInteractiveTest},
		{name: "Poll_Noninteractive", f: poll_NonInteractiveTest},
		{name: "Offer_NonInteractive", f: offer_NonInteractiveTest},
		{name: "Size_Interactive", f: size_InteractiveTest},
		{name: "Size_Noninteractive", f: size_NoninteractiveTest},
		{name: "Destroy_NonInteractive", f: destroy_NonInteractiveTest},
		{name: "Destroy_AutoYes_NonInteractiveTest", f: destroy_autoYes_NonInteractiveTest},
		{name: "Destroy_InteractiveTest", f: destroy_InteractiveTest},
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
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "clear", "-q"))
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
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "poll", "foo", "-q", "--show-type"))
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
			check.MustValue(q.Add(ctx, "foo"))
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

func size_InteractiveTest(t *testing.T) {
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\queue -n %s size\n", q.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/queue_size_0.txt")
			})
			tcx.WithReset(func() {
				check.MustValue(q.Add(ctx, "foo"))
				tcx.WriteStdin([]byte(fmt.Sprintf("\\queue -n %s size\n", q.Name())))
				tcx.AssertStderrContains("OK")
				tcx.AssertStdoutDollarWithPath("testdata/queue_size_1.txt")
			})
		})
	})
}

func destroy_NonInteractiveTest(t *testing.T) {
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			go tcx.WriteStdin([]byte("y\n"))
			check.Must(tcx.CLC().Execute(ctx, "queue", "-n", q.Name(), "destroy"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameQueue, q.Name(), objects))
		})
	})
}

func destroy_autoYes_NonInteractiveTest(t *testing.T) {
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

func destroy_InteractiveTest(t *testing.T) {
	t.Skip()
	it.QueueTester(t, func(tcx it.TestContext, q *hz.Queue) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\queue -n %s destroy\n", q.Name())))
				tcx.AssertStdoutContains("(y/n)")
				tcx.WriteStdin([]byte("y"))
				objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
				require.False(t, objectExists(hz.ServiceNameQueue, q.Name(), objects))
			})
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
