//go:build std || script || alias

package commands_test

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestScript(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "script_Interactive", f: script_InteractiveTest},
		{name: "script_NonInteractive", f: script_NonInteractiveTest},
		{name: "scriptWithAlias_Interactive", f: scriptWithAlias_InteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func script_NonInteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		tcx.CLCExecute(ctx, "script", "testdata/test-script.clc", "--echo", "--ignore-errors")
		tcx.AssertStdoutContains("bar")
		tcx.AssertStderrContains("unknown command")
	})
}

func script_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString(fmt.Sprintf("\\script testdata/test-script.clc --echo --ignore-errors\n"))
				tcx.AssertStdoutContains("bar")
				tcx.AssertStderrContains("unknown command")
			})
		})
	})
}

func scriptWithAlias_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString(fmt.Sprintf("\\alias add mapAlias %s\n", "\\map set key1 value1 -n myMap"))
				tcx.WriteStdinString(fmt.Sprintf("\\script testdata/test-script-with-alias.clc\n"))
				tcx.WriteStdinString(fmt.Sprintf("\\map get key1 -n myMap\n"))
				tcx.AssertStdoutContains("value1")
			})
		})
	})
}
