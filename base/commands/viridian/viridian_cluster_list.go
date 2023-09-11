//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterListCommand struct{}

func (ClusterListCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list-clusters")
	long := `Lists all Viridian clusters for the logged in API key.

Make sure you login before running this command.
`
	short := "Lists Viridian clusters"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (ClusterListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	csi, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Retrieving the clusters")
		cs, err := api.ListClusters(ctx)
		if err != nil {
			return nil, err
		}
		return cs, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	cs := csi.([]viridian.Cluster)
	if len(cs) == 0 {
		ec.PrintlnUnnecessary("OK No clusters found")
		return nil
	}
	rows := make([]output.Row, len(cs))
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	for i, c := range cs {
		rows[i] = output.Row{
			output.Column{
				Name:  "ID",
				Type:  serialization.TypeString,
				Value: c.ID,
			},
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: c.Name,
			},
			output.Column{
				Name:  "State",
				Type:  serialization.TypeString,
				Value: fixClusterState(c.State),
			},
			output.Column{
				Name:  "Hazelcast Version",
				Type:  serialization.TypeString,
				Value: c.HazelcastVersion,
			},
		}
		if verbose {
			rows[i] = append(rows[i],
				output.Column{
					Name:  "Cluster Type",
					Type:  serialization.TypeString,
					Value: ClusterType(c.ClusterType.DevMode),
				},
			)
		}
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:list-clusters", &ClusterListCommand{}))
}
