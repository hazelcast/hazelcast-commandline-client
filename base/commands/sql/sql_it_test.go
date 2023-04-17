package sql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestSQL(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "SQLOutput_NonInteractive", f: sqlOutput_NonInteractiveTest},
		{name: "SQLOutput_JSON", f: sqlOutput_JSONTest},
		{name: "SQLOutput_JSONFlat", f: sqlOutput_JSONFlatTest},
		{name: "SQL_ShellCommand", f: sql_shellCommandTest},
		{name: "SQL_Interactive", f: sql_InteractiveTest},
		{name: "SQL_NonInteractive", f: sql_NonInteractiveTest},
		{name: "SQL_NonInteractiveStreaming", f: sql_NonInteractiveStreamingTest},
		{name: "SQL_Suggestion_Interactive", f: sqlSuggestion_Interactive},
		{name: "SQL_Suggestion_NonInteractive", f: sqlSuggestion_NonInteractive},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func sql_NonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		name := it.NewUniqueObjectName("table")
		ctx := context.Background()
		tcx.CLCExecute(ctx, "sql", fmt.Sprintf(`
			CREATE MAPPING "%s" (
				__key INT,
				this VARCHAR
			) TYPE IMAP OPTIONS (
				'keyFormat' = 'int',
				'valueFormat' = 'varchar'
			);
		`, name))
		tcx.CLCExecute(ctx, "sql", fmt.Sprintf(`
			INSERT INTO "%s" (__key, this) VALUES (10, 'foo'), (20, 'bar');
		`, name))
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "sql", fmt.Sprintf(`
			SELECT * FROM "%s" ORDER BY __key;
		`, name))
			tcx.AssertStdoutContains("10\tfoo\n20\tbar\n")
		})
	})
}

func sql_InteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			name := it.NewUniqueObjectName("table")
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
			tcx.WithReset(func() {
				tcx.WriteStdinf(`SELECT * FROM "%s" ORDER BY __key;`+"\n", name)
				tcx.AssertStdoutDollarWithPath("testdata/sql_1.txt")
			})
		})
	})
}

func sql_shellCommandTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			// help
			tcx.WithReset(func() {
				tcx.WriteStdinString("help\n")
				tcx.AssertStdoutDollarWithPath("testdata/sql_help.txt")
			})
			name := it.NewUniqueObjectName("table")[:16]
			q := fmt.Sprintf(`
				CREATE MAPPING "%s" (
					__key INT,
					this VARCHAR
				) TYPE IMAP OPTIONS (
					'keyFormat' = 'int',
					'valueFormat' = 'varchar'
				);`+"\n", name)
			check.MustValue(tcx.Client.SQL().Execute(ctx, q))
			// dm
			tcx.WithReset(func() {
				tcx.WriteStdinString("\\dm\n")
				tcx.AssertStdoutContains(name)
			})
			// dm NAME
			tcx.WithReset(func() {
				tcx.WriteStdinf("\\dm %s\n", name)
				target := fmt.Sprintf(`$----------------------------------------------------------------------------------------------------$
$ table_catalog | table_schema | table_name | mapping_external_name | mapping_type | mapping_options $
$----------------------------------------------------------------------------------------------------$
$ hazelcast     | public       | test-table | %s      | IMAP         | {"keyFormat":"i $
$----------------------------------------------------------------------------------------------------$`, name)
				tcx.AssertStdoutDollar(target)
			})
			// dm+ NAME
			tcx.WithReset(func() {
				tcx.WriteStdinf("\\dm+ %s\n", name)
				target := `$-----------------------------------------------------------------------------------------------------------------------------$
$ table_catalog | table_schema | table_name | column_name | column_external_name | ordinal_position | is_nullable | data_type $
$-----------------------------------------------------------------------------------------------------------------------------$
$ hazelcast     | public       | test-table | __key       | __key                |                1 | true        | INTEGER   $
$ hazelcast     | public       | test-table | this        | this                 |                2 | true        | VARCHAR   $
$-----------------------------------------------------------------------------------------------------------------------------$`
				tcx.AssertStdoutDollar(target)
			})
		})
	})
}

func sqlSuggestion_Interactive(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		it.WithMap(tcx, func(m *hazelcast.Map) {
			check.Must(m.Set(ctx, "foo", "bar"))
			tcx.WithShell(ctx, func(tcx it.TestContext) {
				tcx.WriteStdinf(`SELECT * FROM "%s";`+"\n", m.Name())
				tcx.AssertStderrContains("CREATE MAPPING")
				tcx.AssertStderrNotContains("--use-mapping-suggestion")
			})
		})
	})
}

func sqlSuggestion_NonInteractive(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		it.WithMap(tcx, func(m *hazelcast.Map) {
			check.Must(m.Set(ctx, "foo", "bar"))
			// ignoring the error here
			_ = tcx.CLC().Execute(ctx, "sql", fmt.Sprintf(`SELECT * FROM "%s";`, m.Name()))
			tcx.AssertStderrContains("CREATE MAPPING")
			tcx.AssertStderrContains("--use-mapping-suggestion")
			check.Must(tcx.CLC().Execute(ctx, "sql", fmt.Sprintf(`SELECT * FROM "%s";`, m.Name()), "--use-mapping-suggestion"))
			tcx.AssertStdoutContains("foo\tbar")
		})
	})
}

func sqlOutput_NonInteractiveTest(t *testing.T) {
	formats := []string{"delimited", "json", "csv", "table"}
	for _, f := range formats {
		f := f
		t.Run(f, func(t *testing.T) {
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				name := it.NewUniqueObjectName("table")
				ctx := context.Background()
				check.MustValue(tcx.Client.SQL().Execute(ctx, fmt.Sprintf(`
			CREATE MAPPING "%s" (
				__key INT,
				this VARCHAR
			) TYPE IMAP OPTIONS (
				'keyFormat' = 'int',
				'valueFormat' = 'varchar'
			);
		`, name)))
				check.MustValue(tcx.Client.SQL().Execute(ctx, fmt.Sprintf(`
			INSERT INTO "%s" (__key, this) VALUES (10, 'foo'), (20, 'bar');
		`, name)))
				tcx.WithReset(func() {
					tcx.CLCExecute(ctx, "sql", "--format", f, "-q", fmt.Sprintf(`
					SELECT * FROM "%s" ORDER BY __key;
				`, name))
					tcx.AssertStdoutDollarWithPath(fmt.Sprintf("testdata/sql_output_%s.txt", f))
				})
			})

		})
	}
}

func sql_NonInteractiveStreamingTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		/*
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
		*/
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "sql", "select * from table(generate_stream(1)) limit 3")
				tcx.AssertStdoutDollar("$1$2$")
			})
		})
	})
}

func sqlOutput_JSONTest(t *testing.T) {
	formats := []string{"delimited", "json", "csv", "table"}
	for _, f := range formats {
		f := f
		t.Run(f, func(t *testing.T) {
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				name := it.NewUniqueObjectName("table")
				ctx := context.Background()
				it.WithEnv(clc.EnvMaxCols, "150", func() {
					check.MustValue(tcx.Client.SQL().Execute(ctx, fmt.Sprintf(`
			CREATE MAPPING "%s"
			TYPE IMAP OPTIONS (
				'keyFormat' = 'int',
				'valueFormat' = 'json'
			);
		`, name)))
					v := `JSON_OBJECT(
			'fieldStr': 'string',
			'fieldNum': 38.27,
			'fieldBool': true,
			'fieldNull': null,
			'fieldArray': JSON_ARRAY(1, 2, 3),
			'fieldObj': JSON_OBJECT('inner': 'foo')
		)`
					check.MustValue(tcx.Client.SQL().Execute(ctx, fmt.Sprintf(`
			INSERT INTO "%s" VALUES (10, %s);
		`, name, v)))
					tcx.WithReset(func() {
						tcx.CLCExecute(ctx, "sql", "--format", f, "-q", fmt.Sprintf(`
					SELECT * FROM "%s" ORDER BY __key;
				`, name))
						tcx.AssertStdoutDollarWithPath(fmt.Sprintf("testdata/sql_output_json_%s.txt", f))
					})
				})
			})
		})
	}
}

func sqlOutput_JSONFlatTest(t *testing.T) {
	formats := []string{"delimited", "json", "csv", "table"}
	for _, f := range formats {
		f := f
		t.Run(f, func(t *testing.T) {
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				name := it.NewUniqueObjectName("table")
				ctx := context.Background()
				it.WithEnv(clc.EnvMaxCols, "150", func() {
					check.MustValue(tcx.Client.SQL().Execute(ctx, fmt.Sprintf(`
			CREATE MAPPING "%s" (
				__key INT,
				fieldStr VARCHAR,
				fieldNum DOUBLE,
				fieldBool BOOLEAN,
				fieldNull VARCHAR
			)
			TYPE IMAP OPTIONS (
				'keyFormat' = 'int',
				'valueFormat' = 'json-flat'
			);
		`, name)))
					v := `'string', 38.27, true, null`
					check.MustValue(tcx.Client.SQL().Execute(ctx, fmt.Sprintf(`
			INSERT INTO "%s"(__key, fieldStr, fieldNum, fieldBool, fieldNull) VALUES (10, %s);
		`, name, v)))
					tcx.WithReset(func() {
						tcx.CLCExecute(ctx, "sql", "--format", f, "-q", fmt.Sprintf(`
					SELECT * FROM "%s" ORDER BY __key;
				`, name))
						tcx.AssertStdoutDollarWithPath(fmt.Sprintf("testdata/sql_output_jsonflat_%s.txt", f))
					})
				})
			})
		})
	}
}
