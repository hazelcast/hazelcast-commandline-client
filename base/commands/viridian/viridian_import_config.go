package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ImportConfigCmd struct{}

func (cmd ImportConfigCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("import-config [cluster-name/cluster-ID] [flags]")
	long := `Imports connection configuration of the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Imports connection configuration of the given Viridian cluster."
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagName, "", "", false, "name of the connection configuration")
	return nil
}

func (cmd ImportConfigCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.Args()[0]
	cluster, err := api.FindCluster(ctx, clusterNameOrID)
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	cfgName := ec.Props().GetString(flagName)
	if cfgName == "" {
		cfgName = cluster.Name
	}
	cp, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Importing configuration")
		zipPath, stop, err := api.DownloadConfig(ctx, cluster.ID)
		if err != nil {
			return nil, err
		}
		defer stop()
		cfgPath, err := config.CreateFromZip(ctx, ec, cfgName, zipPath)
		if err != nil {
			return nil, err
		}
		return cfgPath, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.Logger().Info("Imported configuration %s and saved to: %s", cfgName, cp)
	ec.PrintlnUnnecessary(fmt.Sprintf("Imported configuration: %s", cfgName))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:import-config", &ImportConfigCmd{}))
}
