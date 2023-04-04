package serialization_test

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestBuiltinSerialization(t *testing.T) {
	testCases := []struct {
		name            string
		value           any
		delimitedOutput string
		jsonOutput      string
		csvOutput       string
		tableOutputPath string
	}{
		{
			name:            "types.Decimal",
			value:           types.NewDecimal(big.NewInt(100), 10),
			delimitedOutput: "1.00E-8\n",
			jsonOutput:      `{"this":"1.00E-8"}` + "\n",
			csvOutput:       "this\n1.00E-8\n",
			tableOutputPath: "testdata/builtin_decimal.txt",
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
								check.Must(tcx.CLC().Execute("map", "get", "-n", m.Name(), key, "--quite", "-f", format))
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
									tcx.AssertStdoutDollarWithPath(t, target)
								} else {
									tcx.AssertStdoutEquals(t, target)
								}
							})
						})
					}
				})
			})
		})
	}
}
