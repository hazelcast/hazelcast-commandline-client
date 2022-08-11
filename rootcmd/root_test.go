package rootcmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestNew_HelpContainsFakedoors(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		cmd, _ := New(c)
		var b strings.Builder
		cmd.SetOut(&b)
		cmd.SetArgs([]string{"help"})
		require.NoError(t, cmd.Execute())
		output := b.String()
		require.Contains(t, output, "list")
		require.Contains(t, output, "queue")
		require.Contains(t, output, "map")
		require.Contains(t, output, "multimap")
		require.Contains(t, output, "set")
		require.Contains(t, output, "topic")
		require.Contains(t, output, "replicatedmap")
	})
}

func TestNew(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		type args struct {
			cnfg *hazelcast.Config
		}
		tcs := []struct {
			name     string
			issueNum string
		}{
			{
				name:     "List",
				issueNum: "48",
			},
			{
				name:     "Queue",
				issueNum: "49",
			},
			{
				name:     "MultiMap",
				issueNum: "50",
			},
			{
				name:     "ReplicatedMap",
				issueNum: "51",
			},
			{
				name:     "Set",
				issueNum: "52",
			},
			{
				name:     "Topic",
				issueNum: "53",
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd, _ := New(c)
				cmd.SetArgs([]string{strings.ToLower(tc.name)})
				var sb strings.Builder
				cmd.SetOut(&sb)
				require.NoError(t, cmd.Execute())
				output := sb.String()
				expected := fmt.Sprintf(`The support for %s isn't implemented yet.
Add a thumbs up to it at: https://github.com/hazelcast/hazelcast-commandline-client/issues/%s
`, tc.name, tc.issueNum)
				require.Equal(t, expected, output)
			})
		}
	})
}
