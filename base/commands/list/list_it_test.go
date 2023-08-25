//go:build std || list

package list_test

import (
	"context"
	"fmt"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestList(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Add_NonInteractive", f: add_NonInteractiveTest},
		{name: "Clear_NonInteractive", f: clear_NonInteractiveTest},
		{name: "Contains_NonInteractive", f: contains_NonInteractiveTest},
		{name: "RemoveIndex_Noninteractive", f: removeIndex_NonInteractiveTest},
		{name: "Remove_Noninteractive", f: remove_NonInteractiveTest},
		{name: "Set_NonInteractive", f: set_NonInteractiveTest},
		{name: "Size_Interactive", f: size_InteractiveTest},
		{name: "Size_Noninteractive", f: size_NoninteractiveTest},
		{name: "Destroy_NonInteractive", f: destroy_NonInteractiveTest},
		{name: "Destroy_AutoYes_NonInteractiveTest", f: destroy_autoYes_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func add_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "list", "-n", l.Name(), "add", "foo")
			tcx.CLCExecute(ctx, "list", "-n", l.Name(), "add", "bar")
			require.Equal(t, 0, check.MustValue(l.IndexOf(context.Background(), "foo")))
			require.Equal(t, 1, check.MustValue(l.IndexOf(context.Background(), "bar")))
		})
	})
}

func clear_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			_ = check.MustValue(l.Add(ctx, "foo"))
			require.Equal(t, 1, check.MustValue(l.Size(ctx)))
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "clear", "--yes"))
			require.Equal(t, 0, check.MustValue(l.Size(ctx)))
		})
	})
}

func contains_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "contains", "foo"))
			tcx.AssertStdoutContains("false")
			_ = check.MustValue(l.Add(ctx, "foo"))
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "contains", "foo"))
			tcx.AssertStdoutContains("true")
		})
	})
}

func remove_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		ctx := context.Background()
		tcx.WithReset(func() {
			_ = check.MustValue(l.Add(ctx, "foo"))
			require.Equal(tcx.T, 1, check.MustValue(l.Size(ctx)))
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "remove-value", "foo"))
			require.Equal(tcx.T, 0, check.MustValue(l.Size(ctx)))
		})
	})
}

func removeIndex_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		ctx := context.Background()
		tcx.WithReset(func() {
			_ = check.MustValue(l.Add(ctx, "foo"))
			_ = check.MustValue(l.Add(ctx, "bar"))
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "remove-index", "0"))
			require.Equal(tcx.T, "bar", check.MustValue(l.Get(ctx, 0)))
		})
	})
}

func set_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			_ = check.MustValue(l.Add(ctx, "foo"))
			_ = check.MustValue(l.Add(ctx, "bar"))
			tcx.CLCExecute(ctx, "list", "-n", l.Name(), "set", "0", "foo2")
			index := check.MustValue(l.IndexOf(context.Background(), "foo2"))
			require.Equal(t, 0, index)
		})
	})
}

func size_NoninteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "size"))
			tcx.AssertStdoutEquals("0\n")
		})
		// set an entry
		tcx.WithReset(func() {
			_ = check.MustValue(l.Add(ctx, "foo"))
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "size"))
			tcx.AssertStdoutEquals("1\n")
		})
	})
}

func size_InteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\list -n %s size\n", l.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/list_size_0.txt")
			})
			tcx.WithReset(func() {
				_ = check.MustValue(l.Add(ctx, "foo"))
				tcx.WriteStdin([]byte(fmt.Sprintf("\\list -n %s size\n", l.Name())))
				tcx.AssertStderrContains("OK")
				tcx.AssertStdoutDollarWithPath("testdata/list_size_1.txt")
			})
		})
	})
}

func destroy_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			go tcx.WriteStdin([]byte("y\n"))
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "destroy"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameList, l.Name(), objects))
		})
	})
}

func destroy_autoYes_NonInteractiveTest(t *testing.T) {
	it.ListTester(t, func(tcx it.TestContext, l *hz.List) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "list", "-n", l.Name(), "destroy", "--yes"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameList, l.Name(), objects))
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
