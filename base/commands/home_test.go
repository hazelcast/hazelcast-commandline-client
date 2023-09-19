//go:build std || home

package commands_test

import (
	"context"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it/skip"
)

func TestHome(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "home_unix", f: homeTest_Unix},
		{name: "home_ArgsUnix", f: homeTest_ArgsUnix},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func homeTest_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	ctx := context.Background()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		tcx.CLCExecute(ctx, "home")
		tcx.AssertStdoutEquals(tcx.HomePath() + "\n")
	})
}

func homeTest_ArgsUnix(t *testing.T) {
	skip.If(t, "os = windows")
	ctx := context.Background()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		tcx.CLCExecute(ctx, "home", "foo", "bar")
		tcx.AssertStdoutEquals(tcx.HomePath() + "/foo/bar\n")
	})
}

// TODO: TestHome_Windows
