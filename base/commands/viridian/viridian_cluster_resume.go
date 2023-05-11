package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClusterResumeCmd struct{}

func (cm ClusterResumeCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("resume-cluster [cluster-ID/name] [flags]")
	long := `Resumes the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Resumes the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (cm ClusterResumeCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.Args()[0]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Resuming the cluster")
		err := api.ResumeCluster(ctx, clusterNameOrID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		ec.Logger().Error(err)
		return fmt.Errorf("error resuming the cluster. Did you login?: %w", err)
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:resume-cluster", &ClusterResumeCmd{}))
}
