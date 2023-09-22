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

type ClusterResumeCommand struct{}

func (ClusterResumeCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("resume-cluster")
	long := `Resumes the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Resumes the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	return nil
}

func (ClusterResumeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.Metrics().Increment(metrics.NewSimpleKey(), "total.viridian."+cmd.RunningMode(ec))
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	nameOrID := ec.GetStringArg(argClusterID)
	st := stage.Stage[viridian.Cluster]{
		ProgressMsg: "Starting to resume the cluster",
		SuccessMsg:  "Started to resume the cluster",
		FailureMsg:  "Failed to start resuming the cluster",
		Func: func(ctx context.Context, status stage.Statuser[viridian.Cluster]) (viridian.Cluster, error) {
			return api.ResumeCluster(ctx, nameOrID)
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
	check.Must(plug.Registry.RegisterCommand("viridian:resume-cluster", &ClusterResumeCommand{}))
}
