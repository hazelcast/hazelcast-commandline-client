package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClusterPauseCmd struct{}

func (cm ClusterPauseCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("pause-cluster [cluster-ID/name] [flags]")
	long := `Pauses the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Pauses the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (cm ClusterPauseCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.Args()[0]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Pausing the cluster")
		err := api.StopCluster(ctx, clusterNameOrID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.SetResultString(fmt.Sprintf("Viridian cluster paused: %s", clusterNameOrID))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:pause-cluster", &ClusterPauseCmd{}))
}
