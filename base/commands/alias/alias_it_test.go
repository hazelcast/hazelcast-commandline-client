//go:build std || alias

package alias_test

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands/alias"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestAlias(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Execute_Interactive", f: Execute_InteractiveTest},
		{name: "Add_Interactive", f: Add_InteractiveTest},
		{name: "Remove_Interactive", f: Remove_InteractiveTest},
		{name: "List_Interactive", f: List_InteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func Execute_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		alias.Aliases.Store("mapAlias", "map set 1 1")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString("@mapAlias\n")
				tcx.WriteStdinString("\\map get 1\n")
				tcx.AssertStdoutContains("1")
			})
		})
	})
}

func Add_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString(fmt.Sprintf("\\alias add mapAlias %s\n", `map set 1 1`))
				tcx.WriteStdinString("\\alias list\n")
				tcx.AssertStdoutContains("map set 1 1")
			})
		})
	})
}

func Remove_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		alias.Aliases.Store("mapAlias", "map set 1 1")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString("\\alias remove mapAlias\n")
				tcx.WriteStdinString("\\alias list\n")
				tcx.AssertStdoutNotContains("map set 1 1")
			})
		})
	})
}

func List_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		alias.Aliases.Store("mapAlias", "map set 1 1")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString("\\alias list\n")
				tcx.AssertStdoutContains("mapAlias")
				tcx.AssertStdoutContains("map set 1 1")
			})
		})
	})
}
