package serialization_test

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	pubserialization "github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestBuiltinSerialization(t *testing.T) {
	tv := time.Date(2023, 2, 3, 4, 5, 6, 7, time.UTC)
	testCases := []struct {
		name            string
		value           any
		delimitedOutput string
		jsonOutput      string
		csvOutput       string
	}{
		{
			name:            "byte",
			value:           byte(8),
			delimitedOutput: "8\n",
			jsonOutput:      `{"this":8}` + "\n",
			csvOutput:       "this\n8\n",
		},
		{
			name:            "types.Decimal",
			value:           types.NewDecimal(big.NewInt(100), 10),
			delimitedOutput: "1.00E-8\n",
			jsonOutput:      `{"this":"1.00E-8"}` + "\n",
			csvOutput:       "this\n1.00E-8\n",
		},
		{
			name:            "string",
			value:           "test-string",
			delimitedOutput: "test-string\n",
			jsonOutput:      `{"this":"test-string"}` + "\n",
			csvOutput:       "this\ntest-string\n",
		},
		{
			name:            "types.OffsetDateTime",
			value:           types.OffsetDateTime(tv),
			delimitedOutput: "2023-02-03T04:05:06Z\n",
			jsonOutput:      `{"this":"2023-02-03T04:05:06Z"}` + "\n",
			csvOutput:       "this\n2023-02-03T04:05:06Z\n",
		},
		{
			name:            "serialization.JSON",
			value:           pubserialization.JSON(`{"FieldA":"json-str-1", "FieldB":22}`),
			delimitedOutput: `{"FieldA":"json-str-1", "FieldB":22}` + "\n",
			jsonOutput:      `{"this":{"FieldA":"json-str-1","FieldB":22}}` + "\n",
			csvOutput:       "this\n" + `"{""FieldA"":""json-str-1"", ""FieldB"":22}"` + "\n",
		},
		{
			name:            "slice",
			value:           []any{int64(100), "foo", int32(200), true},
			delimitedOutput: "[100, foo, 200, true]\n",
			jsonOutput:      `{"this":[100,"foo",200,true]}` + "\n",
			csvOutput:       "this\n" + `"[100, foo, 200, true]"` + "\n",
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				it.WithMap(tcx, func(m *hazelcast.Map) {
					key := tc.name
					check.Must(m.Set(ctx, key, tc.value))
					for _, format := range []string{"delimited", "json", "csv", "table"} {
						t.Run(format, func(t *testing.T) {
							tcx.T = t
							tcx.WithReset(func() {
								ctx := context.Background()
								check.Must(tcx.CLC().Execute(ctx, "map", "get", "-n", m.Name(), key, "-q", "-f", format))
								var target string
								switch format {
								case "delimited":
									target = tc.delimitedOutput
								case "json":
									target = tc.jsonOutput
								case "csv":
									target = tc.csvOutput
								case "table":
									target = strings.ToLower(fmt.Sprintf("testdata/builtin_%s.txt", tc.name))
								default:
									panic(fmt.Sprintf("unknown format: %s", format))
								}
								if format == "table" {
									tcx.AssertStdoutDollarWithPath(target)
								} else {
									tcx.AssertStdoutEquals(target)
								}
							})
						})
					}
				})
			})
		})
	}
}
