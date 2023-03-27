package sql_test

import (
	"fmt"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestSQL(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "SQL_Interactive", f: sql_InteractiveTest},
		{name: "SQL_NonInteractive", f: sql_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func sql_NonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		t := tcx.T
		name := it.NewUniqueObjectName("table")
		check.Must(tcx.CLC().Execute("sql", fmt.Sprintf(`
			CREATE MAPPING "%s" (
				__key INT,
				this VARCHAR
			) TYPE IMAP OPTIONS (
				'keyFormat' = 'int',
				'valueFormat' = 'varchar'
			);
		`, name)))
		tcx.AssertStdoutEquals(t, "")
		check.Must(tcx.CLC().Execute("sql", fmt.Sprintf(`
			INSERT INTO "%s" (__key, this) VALUES (10, 'foo'), (20, 'bar');
		`, name)))
		tcx.AssertStdoutEquals(t, "")
		check.Must(tcx.CLC().Execute("sql", fmt.Sprintf(`
			SELECT * FROM "%s" ORDER BY __key;
		`, name)))
		tcx.AssertStdoutEquals(t, "10\tfoo\n20\tbar\n")
	})
}

func sql_InteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		t := tcx.T
		go func(t *testing.T) {
			check.Must(tcx.CLC().Execute())
		}(t)
		name := it.NewUniqueObjectName("table")
		tcx.WriteStdin([]byte(fmt.Sprintf(`
			CREATE MAPPING "%s" (
				__key INT,
				this VARCHAR
			) TYPE IMAP OPTIONS (
				'keyFormat' = 'int',
				'valueFormat' = 'varchar'
			);`+"\n", name)))
		tcx.AssertStdoutEquals(t, "")
		tcx.WriteStdin([]byte(fmt.Sprintf(`
			INSERT INTO "%s" (__key, this) VALUES (10, 'foo'), (20, 'bar');
		`+"\n", name)))
		tcx.AssertStdoutEquals(t, "")
		tcx.WriteStdin([]byte(fmt.Sprintf(`
			SELECT * FROM "%s" ORDER BY __key;
		`+"\n", name)))
		tcx.AssertStdoutContainsWithPath(t, "testdata/sql_1.txt")
	})
}
