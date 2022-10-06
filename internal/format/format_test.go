package format

import (
	"bytes"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

func Test_formatter(t *testing.T) {
	currTime, err := time.Parse(time.RFC3339, "2022-09-01T10:48:32.35+03:00")
	require.NoError(t, err)
	tcs := []struct {
		name     string
		toFormat interface{}
		expected string
	}{
		{
			name:     "LocalDate",
			toFormat: types.LocalDate(currTime),
			expected: "2022-09-01",
		},
		{
			name:     "LocalTime",
			toFormat: types.LocalTime(currTime),
			expected: "10:48:32.35",
		},
		{
			name:     "LocalDateTime",
			toFormat: types.LocalDateTime(currTime),
			expected: "2022-09-01T10:48:32.35",
		},
		{
			name:     "OffsetDateTime",
			toFormat: types.OffsetDateTime(currTime),
			expected: "2022-09-01T10:48:32+03:00",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := Fmt(tc.toFormat)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestWriter(t *testing.T) {
	backup := table.ConsoleSize
	defer func() {
		table.ConsoleSize = backup
	}()
	table.ConsoleSize = func() (int, int) {
		return 40, 100
	}
	tcs := []struct {
		name    string
		format  string
		headers []interface{}
		values  [][]interface{}
		output  string
	}{
		// pretty
		{
			name:    "pretty table with headers and values",
			format:  Pretty,
			headers: []interface{}{"key", "value"},
			values: [][]interface{}{
				{"k1", "v1"},
				{"k2", serialization.JSON(`{"some":"value"}`)},
			},
			output: "+-------------------------------------+\n" +
				"|        key       |       value      |\n" +
				"+-------------------------------------+\n" +
				"| k1               | v1               |\n" +
				`| k2               | {"some":"value"} |`,
		},
		{
			name:   "pretty table with no headers",
			format: Pretty,
			values: [][]interface{}{
				{"k1", "v1"},
				{"k2", serialization.JSON(`{"some":"value"}`)},
			},
			output: "| k1               | v1               |\n" +
				`| k2               | {"some":"value"} |`,
		},
		// CSV
		{
			name:    "csv table with headers and values",
			format:  CSV,
			headers: []interface{}{"key", "value"},
			values: [][]interface{}{
				{"k1", "v1"},
				{"k2", serialization.JSON(`{"some":"value"}`)},
			},
			// extra quotes are due to CSV format
			output: `key,value
k1,v1
k2,"{""some"":""value""}"`,
		},
		{
			name:   "csv table with no headers",
			format: CSV,
			values: [][]interface{}{
				{"k1", "v1"},
				{"k2", serialization.JSON(`{"some":"value"}`)},
			},
			// extra quotes are due to CSV format
			output: `k1,v1
k2,"{""some"":""value""}"`,
		},
		// JSON
		{
			name:    "json table with headers and values",
			format:  JSON,
			headers: []interface{}{"key", "value"},
			values: [][]interface{}{
				{"k1", "v1"},
				{"k2", serialization.JSON(`{"some":"value"}`)},
			},
			// extra quotes are due to CSV format
			output: `{"key":"k1","value":"v1"}
{"key":"k2","value":{"some":"value"}}`,
		},
		{
			name:   "json table with no headers",
			format: JSON,
			values: [][]interface{}{
				{"k1", "v1"},
				{"k2", serialization.JSON(`{"some":"value"}`)},
			},
			// extra quotes are due to CSV format
			output: `{"0":"k1","1":"v1"}
{"0":"k2","1":{"some":"value"}}`,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var bb bytes.Buffer
			w, err := NewWriterBuilder().
				WithOut(&bb).
				WithFormat(tc.format).
				WithHeaders(tc.headers...).Build()
			require.NoError(t, err)
			for _, values := range tc.values {
				require.NoError(t, w(values...))
			}
			require.Equal(t, tc.output+"\n", bb.String())
		})
	}
}
