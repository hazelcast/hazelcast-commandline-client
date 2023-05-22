package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
)

type ClusterDeleteCmd struct{}

func (cm ClusterDeleteCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("delete-cluster [cluster-ID/name] [flags]")
	long := `Deletes the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Deletes the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the delete operation")
	return nil
}

func (cm ClusterDeleteCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Cluster will be deleted irreversibly, proceed?")
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	clusterNameOrID := ec.Args()[0]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Deleting the cluster")
		err := api.DeleteCluster(ctx, clusterNameOrID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.PrintlnUnnecessary(fmt.Sprintf("Cluster %s was deleted.", clusterNameOrID))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:delete-cluster", &ClusterDeleteCmd{}))
}
