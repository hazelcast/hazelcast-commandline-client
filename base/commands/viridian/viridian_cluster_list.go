package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterListCmd struct{}

func (cm ClusterListCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("cluster-list")
	help := "List Viridian clusters"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm ClusterListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	tokenPaths, err := secrets.FindAll(secretPrefix)
	if err != nil {
		return fmt.Errorf("cannot access the secrets, did you login?: %w", err)
	}
	if len(tokenPaths) == 0 {
		return fmt.Errorf("no secrets found, did you login?")
	}
	tp := tokenPaths[0]
	ec.Logger().Info("Using Viridian secret at: %s", tp)
	token, err := secrets.Read(secretPrefix, tp)
	if err != nil {
		ec.Logger().Error(err)
		return fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	api := viridian.NewAPI(token)
	csi, cancel, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Retrieving clusters")
		cs, err := api.ListClusters(ctx)
		if err != nil {
			return nil, err
		}
		return cs, nil
	})
	if err != nil {
		ec.Logger().Error(err)
		return fmt.Errorf("error getting clusters. Did you login?: %w", err)
	}
	cancel()
	cs := csi.([]viridian.Cluster)
	rows := make([]output.Row, len(cs))
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
				Name:  "Hazelcast Version",
				Type:  serialization.TypeString,
				Value: c.HazelcastVersion,
			},
		}
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:cluster-list", &ClusterListCmd{}))
}
