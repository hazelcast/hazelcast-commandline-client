//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterStopCommand struct{}

func (ClusterStopCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("stop-cluster")
	long := `Stops the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Stops the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	return nil
}

func (ClusterStopCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.Metrics().Increment(metrics.NewSimpleKey(), "total.viridian."+cmd.RunningModeString(ec))
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	nameOrID := ec.GetStringArg(argClusterID)
	st := stage.Stage[viridian.Cluster]{
		ProgressMsg: "Initiating cluster stop",
		SuccessMsg:  "Initiated cluster stop",
		FailureMsg:  "Failed to initiate cluster stop",
		Func: func(ctx context.Context, status stage.Statuser[viridian.Cluster]) (viridian.Cluster, error) {
			return api.StopCluster(ctx, nameOrID)
		},
	}
	cluster, err := stage.Execute(ctx, ec, viridian.Cluster{}, stage.NewFixedProvider(st))
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "ID",
			Type:  serialization.TypeString,
			Value: cluster.ID,
		},
	})
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:stop-cluster", &ClusterStopCommand{}))
}
