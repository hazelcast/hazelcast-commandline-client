//go:build std || viridian

package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ImportConfigCommand struct{}

func (cm ImportConfigCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("import-config")
	long := `Imports connection configuration of the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Imports connection configuration of the given Viridian cluster."
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagName, "", "", false, "name of the connection configuration")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	return nil
}

func (cm ImportConfigCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.Metrics().Increment(metric.NewSimpleKey(), "total.viridian."+cmd.RunningMode(ec))
	if err := cm.exec(ctx, ec); err != nil {
		err = handleErrorResponse(ec, err)
		return fmt.Errorf("could not import cluster configuration: %w", err)
	}
	return nil
}

func (cm ImportConfigCommand) exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.GetStringArg(argClusterID)
	c, err := api.FindCluster(ctx, clusterNameOrID)
	if err != nil {
		return err
	}
	cfgName := ec.Props().GetString(flagName)
	if cfgName == "" {
		cfgName = c.Name
	}
	st := stage.Stage[string]{
		ProgressMsg: "Importing the configuration",
		SuccessMsg:  "Imported the configuration",
		FailureMsg:  "Failed importing the configuration",
		Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
			return tryImportConfig(ctx, ec, api, c.ID, cfgName)
		},
	}
	path, err := stage.Execute(ctx, ec, "", stage.NewFixedProvider(st))
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Configuration Path",
			Type:  iserialization.TypeString,
			Value: path,
		},
	})
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:import-config", &ImportConfigCommand{}))
}
