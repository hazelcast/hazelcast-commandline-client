package sqlcmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"
	console "github.com/nathan-fiscaletti/consolesize-go"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

func TestSQLCmd(t *testing.T) {
	// set consoleSize for proper SQL table output
	table.ConsoleSize = func() (int, int) {
		return 100, 100
	}
	defer func() {
		table.ConsoleSize = console.GetConsoleSize
	}()
	const executeQueryOutput = `---
Affected rows: 0

`
	it.SQLTester(t, func(t *testing.T, client *hz.Client, config *hz.Config, m *hz.Map, mapName string) {
		var (
			mappingQry = fmt.Sprintf(`
				CREATE MAPPING "%s" (
					__key INT,
					countries VARCHAR,
					cities VARCHAR)
				TYPE IMap
				OPTIONS('keyFormat'='int', 'valueFormat'='json-flat');`, mapName)
			insertQry = fmt.Sprintf(`
				INSERT INTO "%s" VALUES
				(1, 'United Kingdom','London'),
				(2, 'United Kingdom','Manchester'),
				(3, 'United States', 'New York'),
				(4, 'United States', 'Los Angeles'),
				(5, 'Turkey', 'Ankara'),
				(6, 'Turkey', 'Istanbul'),
				(7, 'Brazil', 'Sao Paulo'),
				(8, 'Brazil', 'Rio de Janeiro')`, mapName)
			selectQry = fmt.Sprintf(`SELECT __key, this from "%s" order by __key`, mapName)
		)
		// case order matters for this test
		tcs := []struct {
			name   string
			args   []string
			output string
			err    error
		}{
			{
				name:   "valid create mapping query",
				args:   []string{mappingQry},
				output: executeQueryOutput,
			},
			{
				name:   "valid insert query",
				args:   []string{insertQry},
				output: executeQueryOutput,
			},
			{
				name: "valid select query",
				args: []string{selectQry},
				output: `+-------------------------------------------------------------------------------------------------+
|                      __key                     |                      this                      |
+-------------------------------------------------------------------------------------------------+
| 1                                              | {"countries":"United Kingdom","cities":"Lon... |
| 2                                              | {"countries":"United Kingdom","cities":"Man... |
| 3                                              | {"countries":"United States","cities":"New ... |
| 4                                              | {"countries":"United States","cities":"Los ... |
| 5                                              | {"countries":"Turkey","cities":"Ankara"}       |
| 6                                              | {"countries":"Turkey","cities":"Istanbul"}     |
| 7                                              | {"countries":"Brazil","cities":"Sao Paulo"}    |
| 8                                              | {"countries":"Brazil","cities":"Rio de Jane... |
`,
			},
			{
				name: "valid select query with csv output",
				args: []string{selectQry, "--output-format", "csv"},
				output: `__key,this
1,"{""countries"":""United Kingdom"",""cities"":""London""}"
2,"{""countries"":""United Kingdom"",""cities"":""Manchester""}"
3,"{""countries"":""United States"",""cities"":""New York""}"
4,"{""countries"":""United States"",""cities"":""Los Angeles""}"
5,"{""countries"":""Turkey"",""cities"":""Ankara""}"
6,"{""countries"":""Turkey"",""cities"":""Istanbul""}"
7,"{""countries"":""Brazil"",""cities"":""Sao Paulo""}"
8,"{""countries"":""Brazil"",""cities"":""Rio de Janeiro""}"
`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := New(config)
				var b bytes.Buffer
				cmd.SetOut(&b)
				ctx := context.Background()
				cmd.SetArgs(tc.args)
				_, err := cmd.ExecuteContextC(ctx)
				if tc.err != nil {
					require.Equal(t, tc.err, err)
				}
				require.Nil(t, err)
				require.Equal(t, tc.output, b.String())
			})
		}
	})
}

func TestSQL_CancelContext(t *testing.T) {
	it.SQLTester(t, func(t *testing.T, client *hz.Client, config *hz.Config, m *hz.Map, mapName string) {
		cmd := New(config)
		//b := &strings.Builder{}
		r, w := io.Pipe()
		cmd.SetOut(w)
		cmd.SetArgs([]string{"select * from table(generate_stream(1))"})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// read the first line to make sure SQL fetch page is called, and we do not
		// cancel the call prematurely.
		go func() {
			s := bufio.NewScanner(r)
			if !s.Scan() {
				panic("can not read SQL command result")
			}
			t.Log(s.Text())
			cancel()
			// keep reading not to block the cmd
			for s.Scan() {
			}
		}()
		_, err := cmd.ExecuteContextC(ctx)
		require.NoError(t, err)
		// release the go routine
		require.NoError(t, w.Close())
	})
}
