package listmemberscmd_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/listmemberscmd"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestMembers(t *testing.T) {
	tc := it.StartNewClusterWithOptions(t.Name(), it.NextPort(), 5)
	defer tc.Shutdown()
	dc := tc.DefaultConfig()
	tcs := []struct {
		name   string
		args   []string
		expect []string
	}{
		{
			name: "list-members",
			args: []string{},
			expect: []string{
				fmt.Sprintf("0"),
			},
		},
	}
	for _, tc := range tcs {
		var b bytes.Buffer
		cmd := listmemberscmd.New(&dc)
		cmd.SetOut(&b)
		ctx := context.Background()
		cmd.SetArgs(tc.args)
		_, err := cmd.ExecuteContextC(ctx)
		require.NoError(t, err)
		out := strings.Split(strings.TrimSpace(b.String()), "\n")
		require.Equal(t, tc.expect, out)
	}
}
