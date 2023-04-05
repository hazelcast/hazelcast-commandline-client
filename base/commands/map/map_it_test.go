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
		{name: "Clear_NonInteractive", f: clear_NonInteractiveTest},
		{name: "EntrySet_NonInteractive", f: entrySet_NonInteractiveTest},
		{name: "Get_Noninteractive", f: get_NonInteractiveTest},
		{name: "Set_NonInteractive", f: set_NonInteractiveTest},
		{name: "Size_Interactive", f: size_InteractiveTest},
		{name: "Size_Noninteractive", f: size_NoninteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func clear_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(m.Set(ctx, "foo", "bar"))
			require.Equal(t, 1, check.MustValue(m.Size(ctx)))
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "clear", "--quite"))
			require.Equal(t, 0, check.MustValue(m.Size(ctx)))
		})
	})

}

func entrySet_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "entry-set", "--quite"))
			tcx.AssertStdoutEquals("")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(m.Set(context.Background(), "foo", "bar"))
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "entry-set", "--quite"))
			tcx.AssertStdoutContains("foo\tbar\n")
		})
		// show type
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "entry-set", "--show-type", "--quite"))
			tcx.AssertStdoutContains("foo\tSTRING\tbar\tSTRING\n")
		})
	})
}

func get_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "foo", "--quite"))
			tcx.AssertStdoutEquals("-\n")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(m.Set(context.Background(), "foo", "bar"))
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "foo", "--quite"))
			tcx.AssertStdoutContains("bar\n")
		})
	})
}

func set_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		tcx.WithReset(func() {
			tcx.CLCExecute("map", "-n", m.Name(), "set", "foo", "bar", "--quite")
			tcx.AssertStderrEquals("")
			v := check.MustValue(m.Get(context.Background(), "foo"))
			require.Equal(t, "bar", v)
		})
	})
}

func size_NoninteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "size", "--quite"))
			tcx.AssertStdoutEquals("0\n")
		})
		// set an entry
		tcx.WithReset(func() {
			tcx.AssertStdoutEquals("")
			check.Must(m.Set(context.Background(), "foo", "bar"))
			check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "size", "--quite"))
			tcx.AssertStdoutEquals("1\n")
		})
	})
}

func size_InteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		tcx.WithShell(func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/map_size_0.txt")
			})
			tcx.WithReset(func() {
				check.Must(m.Set(ctx, "foo", "bar"))
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/map_size_1.txt")
			})
		})
	})
}
