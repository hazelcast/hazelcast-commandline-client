package _map_test

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

func TestMap(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "EntrySet_NonInteractive", f: entrySet_NonInteractiveTest},
		{name: "Get_Noninteractive", f: get_NonInteractiveTest},
		{name: "Set_NonInteractive", f: set_NonInteractiveTest},
		{name: "Size_Noninteractive", f: size_NoninteractiveTest},
		{name: "Size_Interactive", f: size_InteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func entrySet_NonInteractiveTest(t *testing.T) {
	mapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		// no entry
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "entry-set"))
		tcx.AssertStdoutEquals(t, "")
		// set an entry
		check.Must(m.Set(context.Background(), "foo", "bar"))
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "entry-set"))
		tcx.AssertStdoutEquals(t, "foo\tbar\n")
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "entry-set", "--show-type"))
		tcx.AssertStdoutEquals(t, "foo\tSTRING\tbar\tSTRING\n")
	})
}

func get_NonInteractiveTest(t *testing.T) {
	mapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		// no entry
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "foo"))
		tcx.AssertStdoutEquals(t, "-\n")
		// set an entry
		check.Must(m.Set(context.Background(), "foo", "bar"))
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "foo"))
		tcx.AssertStdoutEquals(t, "bar\n")
	})
}

func set_NonInteractiveTest(t *testing.T) {
	mapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "set", "foo", "bar"))
		tcx.AssertStdoutEquals(t, "")
		v := check.MustValue(m.Get(context.Background(), "foo"))
		require.Equal(t, "bar", v)
	})
}

func size_NoninteractiveTest(t *testing.T) {
	mapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		// no entry
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "size"))
		tcx.AssertStdoutEquals(t, "0\n")
		// set an entry
		check.Must(m.Set(context.Background(), "foo", "bar"))
		check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "size"))
		tcx.AssertStdoutEquals(t, "1\n")
	})
}

func size_InteractiveTest(t *testing.T) {
	mapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		go func(t *testing.T) {
			check.Must(tcx.CLC().Execute())
		}(t)
		tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
		tcx.AssertStdoutContainsWithPath(t, "testdata/map_size_0.txt")
		check.Must(m.Set(ctx, "foo", "bar"))
		tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
		tcx.AssertStdoutContainsWithPath(t, "testdata/map_size_1.txt")
	})
}

func withMap(tcx it.TestContext, fn func(m *hz.Map)) {
	name := it.NewUniqueObjectName("map")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetMap(ctx, name))
	fn(m)
}

func mapTester(t *testing.T, fn func(tcx it.TestContext, m *hz.Map)) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		withMap(tcx, func(m *hz.Map) {
			fn(tcx, m)
		})
	})
}
