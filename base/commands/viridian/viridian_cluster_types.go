package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterTypeListCmd struct{}

func (ct ClusterTypeListCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list-cluster-types [flags]")
	long := `Lists available cluster types that can be used while creating a Viridian cluster.

Make sure you login before running this command.
`
	short := "Lists Viridian cluster types"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 0)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (ct ClusterTypeListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	csi, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Retrieving cluster types")
		cs, err := api.ListClusterTypes(ctx)
		if err != nil {
			return nil, err
		}
		return cs, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	cs := csi.([]viridian.ClusterType)
	var rows []output.Row
	for _, c := range cs {
		var r output.Row
		if verbose {
			r = append(r, output.Column{
				Name:  "ID",
				Type:  serialization.TypeInt64,
				Value: c.ID,
			})
		}
		r = append(r, output.Column{
			Name:  "Name",
			Type:  serialization.TypeString,
			Value: c.Name,
		})
		rows = append(rows, r)
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:list-cluster-types", &ClusterTypeListCmd{}))
}
