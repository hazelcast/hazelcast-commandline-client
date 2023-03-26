package _map_test

import (
	"context"
	"fmt"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/require"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestMapSizeCommand_Noninteractive(t *testing.T) {
	tcx := it.TestContext{
		T: t,
	}
	tcx.Tester(func(tcx it.TestContext) {
		withMap(tcx, func(m *hz.Map) {
			t := tcx.T
			// no key set
			require.NoError(t, tcx.CLC().Execute("map", "-n", m.Name(), "size"))
			require.Equal(t, "0\n", string(tcx.ReadStdout()))
			require.Equal(t, "", string(tcx.ReadStderr()))
			// set the first key
			check.Must(m.Set(context.Background(), "foo", "bar"))
			require.NoError(t, tcx.CLC().Execute("map", "-n", m.Name(), "size"))
			require.Equal(t, "1\n", string(tcx.ReadStdout()))
			require.Equal(t, "", string(tcx.ReadStderr()))
			// set the second key
			check.Must(m.Set(context.Background(), "zoo", "quux"))
			require.NoError(t, tcx.CLC().Execute("map", "-n", m.Name(), "size"))
			require.Equal(t, "2\n", string(tcx.ReadStdout()))
			require.Equal(t, "", string(tcx.ReadStderr()))
		})
	})
}

func TestMapSizeCommand_Interactive(t *testing.T) {
	tcx := it.TestContext{
		T: t,
	}
	tcx.Tester(func(tcx it.TestContext) {
		withMap(tcx, func(m *hz.Map) {
			t := tcx.T
			ctx := context.Background()
			go func(t *testing.T) {
				require.NoError(t, tcx.CLC().Execute())
			}(t)
			tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
			tcx.AssertStdoutContainsWithPath(t, "testdata/map_size_0.txt")
			check.Must(m.Set(ctx, "foo", "bar"))
			tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
			tcx.AssertStdoutContainsWithPath(t, "testdata/map_size_1.txt")
			tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size -f json\n", m.Name())))
		})
	})
}

func withMap(tcx it.TestContext, fn func(m *hz.Map)) {
	name := it.NewUniqueObjectName("map")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetMap(ctx, name))
	fn(m)
}
