//go:build base || set

package set_test

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

func TestSet(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Add_NonInteractive", f: add_NonInteractiveTest},
		{name: "Remove_Noninteractive", f: remove_NonInteractiveTest},
		{name: "Clear_NonInteractive", f: clear_NonInteractiveTest},
		{name: "Size_Interactive", f: size_InteractiveTest},
		{name: "Destroy_NonInteractive", f: destroy_NonInteractiveTest},
		{name: "GetAll_NonInteractive", f: getAll_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func getAll_NonInteractiveTest(t *testing.T) {
	it.SetTester(t, func(tcx it.TestContext, s *hz.Set) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			added, err := s.Add(ctx, "foo")
			require.Equal(t, nil, err)
			require.Equal(t, true, added)
			added, err = s.Add(ctx, "foo2")
			require.Equal(t, nil, err)
			require.Equal(t, true, added)
			check.Must(tcx.CLC().Execute(ctx, "set", "-n", s.Name(), "get-all", "-q"))
			tcx.AssertStdoutContains("foo")
			tcx.AssertStdoutContains("foo2")
		})
	})
}

func destroy_NonInteractiveTest(t *testing.T) {
	it.SetTester(t, func(tcx it.TestContext, s *hz.Set) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			done, err := s.Add(ctx, "foo")
			require.Equal(t, nil, err)
			require.Equal(t, true, done)
			check.Must(tcx.CLC().Execute(ctx, "set", "-n", s.Name(), "destroy", "--yes"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameSet, s.Name(), objects))
		})
	})
}

func size_InteractiveTest(t *testing.T) {
	it.SetTester(t, func(tcx it.TestContext, s *hz.Set) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			done, err := s.Add(ctx, "foo")
			require.Equal(t, nil, err)
			require.Equal(t, true, done)
			check.Must(tcx.CLC().Execute(ctx, "set", "-n", s.Name(), "size", "-q"))
			tcx.AssertStdoutEquals("1\n")
		})
	})
}

func clear_NonInteractiveTest(t *testing.T) {
	it.SetTester(t, func(tcx it.TestContext, s *hz.Set) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			done, err := s.Add(ctx, "foo")
			require.Equal(t, nil, err)
			require.Equal(t, true, done)
			require.Equal(t, 1, check.MustValue(s.Size(ctx)))
			check.Must(tcx.CLC().Execute(ctx, "set", "-n", s.Name(), "clear", "--yes", "-q"))
			require.Equal(t, 0, check.MustValue(s.Size(ctx)))
		})
	})
}

func remove_NonInteractiveTest(t *testing.T) {
	it.SetTester(t, func(tcx it.TestContext, s *hz.Set) {
		ctx := context.Background()
		tcx.WithReset(func() {
			added, err := s.Add(ctx, "foo")
			require.Equal(t, nil, err)
			require.Equal(t, true, added)
			require.Equal(tcx.T, 1, check.MustValue(s.Size(ctx)))
			check.Must(tcx.CLC().Execute(ctx, "set", "-n", s.Name(), "remove", "foo", "-q", "--show-type"))
			tcx.AssertStdoutEquals("true\tBOOL\n")
			require.Equal(tcx.T, 0, check.MustValue(s.Size(ctx)))
		})
	})
}

func add_NonInteractiveTest(t *testing.T) {
	it.SetTester(t, func(tcx it.TestContext, s *hz.Set) {
		ctx := context.Background()
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "set", "-n", s.Name(), "add", "foo", "-q")
			tcx.AssertStderrEquals("")
			v := check.MustValue(s.GetAll(ctx))
			require.Contains(t, v, "foo")
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
