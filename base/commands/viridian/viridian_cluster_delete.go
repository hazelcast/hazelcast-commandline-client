//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterDeleteCmd struct{}

func (cm ClusterDeleteCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("delete-cluster")
	long := `Deletes the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Deletes the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the delete operation")
	cc.AddStringArg(argClusterID, argTitleClusterID)
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
	nameOrID := ec.GetStringArg(argClusterID)
	st := stage.Stage[viridian.Cluster]{
		ProgressMsg: "Initiating cluster deletion",
		SuccessMsg:  "Inititated cluster deletion",
		FailureMsg:  "Failed to inititate cluster deletion",
		Func: func(ctx context.Context, status stage.Statuser[viridian.Cluster]) (viridian.Cluster, error) {
			cluster, err := api.DeleteCluster(ctx, nameOrID)
			if err != nil {
				return cluster, err
			}
			return cluster, nil
		},
	}
	cluster, err := stage.Execute(ctx, ec, viridian.Cluster{}, stage.NewFixedProvider(st))
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	ec.PrintlnUnnecessary("")
	row := []output.Column{
		{
			Name:  "ID",
			Type:  serialization.TypeString,
			Value: cluster.ID,
		},
	}
	if ec.Props().GetBool(clc.PropertyVerbose) {
		row = append(row, output.Column{
			Name:  "ID",
			Type:  serialization.TypeString,
			Value: cluster.ID,
		})
	}
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:delete-cluster", &ClusterDeleteCmd{}))
}
