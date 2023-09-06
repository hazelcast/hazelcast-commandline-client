//go:build std || viridian

package viridian

import (
	"context"
	"fmt"

	hzerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ImportConfigCmd struct{}

func (ImportConfigCmd) Init(cc plug.InitContext) error {
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

func (cm ImportConfigCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	if err := cm.exec(ctx, ec); err != nil {
		ec.PrintlnUnnecessary(fmt.Sprintf("FAIL Could not import cluster configuration: %s", err.Error()))
		return hzerrors.WrappedError{Err: err}
	}
	return nil
}

func (ImportConfigCmd) exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.GetStringArg(argClusterID)
	c, err := api.FindCluster(ctx, clusterNameOrID)
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	cfgName := ec.Props().GetString(flagName)
	if cfgName == "" {
		cfgName = c.Name
	}
	if _, err = tryImportConfig(ctx, ec, api, c.ID, cfgName); err != nil {
		return handleErrorResponse(ec, err)
	}
	return nil
}

func (ImportConfigCmd) Unwrappable() {}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:import-config", &ImportConfigCmd{}))
}
