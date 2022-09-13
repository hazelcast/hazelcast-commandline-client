package list_objects

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestList(t *testing.T) {
	tc := it.StartNewClusterWithOptions(t.Name(), it.NextPort(), 1)
	defer tc.Shutdown()
	ctx := context.Background()
	dc := tc.DefaultConfig()
	c, err := hazelcast.StartNewClientWithConfig(ctx, dc)
	require.NoError(t, err)
	// create two to check ordering of the same type
	mapName1, mapName2 := fmt.Sprintf("%s-map1", t.Name()), fmt.Sprintf("%s-map2", t.Name())
	_, err = c.GetMap(ctx, mapName1)
	require.NoError(t, err)
	_, err = c.GetMap(ctx, mapName2)
	require.NoError(t, err)
	// PN Counter
	pnCounterName := it.NewUniqueObjectName("pnCounter")
	_, err = c.GetPNCounter(ctx, pnCounterName)
	require.NoError(t, err)
	// List
	listName := it.NewUniqueObjectName("list")
	_, err = c.GetList(ctx, listName)
	require.NoError(t, err)
	// MultiMap
	mmapName := it.NewUniqueObjectName("multimap")
	_, err = c.GetMultiMap(ctx, mmapName)
	require.NoError(t, err)
	// Queue
	queueName := it.NewUniqueObjectName("queue")
	_, err = c.GetQueue(ctx, queueName)
	require.NoError(t, err)
	// FlakeIDGenerator
	figName := it.NewUniqueObjectName("flakeIDGen")
	_, err = c.GetFlakeIDGenerator(ctx, figName)
	require.NoError(t, err)
	// Replicated Map
	repMapName := it.NewUniqueObjectName("replicatedMap")
	_, err = c.GetReplicatedMap(ctx, repMapName)
	require.NoError(t, err)
	// Set
	setName := it.NewUniqueObjectName("set")
	_, err = c.GetSet(ctx, setName)
	require.NoError(t, err)
	// Topic
	topicName := it.NewUniqueObjectName("topic")
	_, err = c.GetReplicatedMap(ctx, topicName)
	require.NoError(t, err)
	tcs := []struct {
		name   string
		args   []string
		expect []string
	}{
		{
			name: "list all",
			args: []string{},
			expect: []string{
				// todo find out why flake id generators output as "PNCounter"
				fmt.Sprintf("PNCounter %s", figName),
				fmt.Sprintf("PNCounter %s", pnCounterName),
				fmt.Sprintf("list %s", listName),
				fmt.Sprintf("map %s", mapName1),
				fmt.Sprintf("map %s", mapName2),
				fmt.Sprintf("multiMap %s", mmapName),
				fmt.Sprintf("queue %s", queueName),
				fmt.Sprintf("replicatedMap %s", repMapName),
				// todo find out why topics output as "replicatedMap"
				fmt.Sprintf("replicatedMap %s", topicName),
				fmt.Sprintf("set %s", setName),
			},
		},
		{
			name: "list map",
			args: []string{"--type", "map"},
			expect: []string{
				fmt.Sprintf("%s", mapName1),
				fmt.Sprintf("%s", mapName2),
			},
		},
	}
	for _, tc := range tcs {
		var b bytes.Buffer
		cmd := New(&dc)
		cmd.SetOut(&b)
		ctx := context.Background()
		cmd.SetArgs(tc.args)
		_, err := cmd.ExecuteContextC(ctx)
		require.NoError(t, err)
		out := strings.Split(strings.TrimSpace(b.String()), "\n")
		// at the of writing this, hazelcast creates an internal replicated map
		// named "__sql.catalog". The line belows removes that
		out = filterOutInternalObjects(out)
		require.Equal(t, tc.expect, out)
	}
}

func filterOutInternalObjects(out []string) []string {
	for i, l := range out {
		if strings.Contains(l, "__") {
			out = append(out[:i], out[i+1:]...)
		}
	}
	return out
}
