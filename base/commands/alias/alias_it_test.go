//go:build std || alias

package alias_test

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands/alias"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestAlias(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Execute_Interactive", f: Execute_InteractiveTest},
		{name: "ExecuteSQL_Interactive", f: ExecuteSQL_InteractiveTest},
		{name: "Add_Interactive", f: Add_InteractiveTest},
		{name: "Remove_Interactive", f: Remove_InteractiveTest},
		{name: "List_Interactive", f: List_InteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func Execute_InteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		alias.Aliases.Store("mapAlias", fmt.Sprintf("\\map set key1 value1 -n myMap"))
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString("@mapAlias\n")
				tcx.WriteStdinString("\\map get key1 -n myMap\n")
				tcx.AssertStdoutContains("key1")
				tcx.AssertStdoutContains("value1")
			})
		})
	})
}

func ExecuteSQL_InteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		name := it.NewUniqueObjectName("table")
		alias.Aliases.Store("sqlAlias", fmt.Sprintf(`SELECT * FROM "%s" ORDER BY __key;`+"\n", name))
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinf(`
				CREATE MAPPING "%s" (
					__key INT,
					this VARCHAR
				) TYPE IMAP OPTIONS (
					'keyFormat' = 'int',
					'valueFormat' = 'varchar'
				);`+"\n", name)
				tcx.WriteStdinf(`
				INSERT INTO "%s" (__key, this) VALUES (10, 'foo'), (20, 'bar');
			`+"\n", name)
				tcx.WriteStdinString("@sqlAlias\n")
				tcx.AssertStdoutContains("10")
				tcx.AssertStdoutContains("foo")
				tcx.AssertStdoutContains("20")
				tcx.AssertStdoutContains("bar")
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
				tcx.WriteStdinString(fmt.Sprintf("\\alias add mapAlias %s\n", `map set key1 value1`))
				tcx.WriteStdinString("\\alias list\n")
				tcx.AssertStdoutContains("map set key1 value1")
			})
		})
	})
}

func Remove_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		alias.Aliases.Store("mapAlias", "map set key1 value1")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString("\\alias remove mapAlias\n")
				tcx.WriteStdinString("\\alias list\n")
				tcx.AssertStdoutNotContains("map set key1 value1")
			})
		})
	})
}

func List_InteractiveTest(t *testing.T) {
	ctx := context.TODO()
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		alias.Aliases.Store("mapAlias", "map set key1 value1")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString("\\alias list\n")
				tcx.AssertStdoutContains("mapAlias")
				tcx.AssertStdoutContains("map set key1 value1")
			})
		})
	})
}
