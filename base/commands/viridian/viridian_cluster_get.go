//go:build std || viridian

package viridian

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterGetCmd struct{}

func (cm ClusterGetCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("get-cluster [cluster-ID/name] [flags]")
	long := `Gets the information about the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Gets the information about the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (cm ClusterGetCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.Args()[0]
	ci, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Retrieving the cluster")
		c, err := api.GetCluster(ctx, clusterNameOrID)
		if err != nil {
			return nil, err
		}
		return c, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	c := ci.(viridian.Cluster)
	row := output.Row{
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
	if ec.Props().GetBool(clc.PropertyVerbose) {
		row = append(row,
			output.Column{
				Name:  "Creation Time",
				Type:  serialization.TypeJavaLocalDateTime,
				Value: time.UnixMilli(c.CreationTime),
			},
			output.Column{
				Name:  "Start Time",
				Type:  serialization.TypeJavaLocalDateTime,
				Value: time.UnixMilli(c.StartTime),
			},
			output.Column{
				Name:  "Hot Backup Enabled",
				Type:  serialization.TypeString,
				Value: boolToYesNo(c.HotBackupEnabled),
			},
			output.Column{
				Name:  "Hot Restart Enabled",
				Type:  serialization.TypeString,
				Value: boolToYesNo(c.HotRestartEnabled),
			},
			output.Column{
				Name:  "IP Whitelist Enabled",
				Type:  serialization.TypeString,
				Value: boolToYesNo(c.IPWhitelistEnabled),
			},
			output.Column{
				Name:  "Regions",
				Type:  serialization.TypeStringArray,
				Value: regionTitleSlice(c.Regions),
			},
		)
	}
	return ec.AddOutputRows(ctx, row)
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func regionTitleSlice(regions []viridian.Region) []string {
	titles := []string{}
	for _, region := range regions {
		titles = append(titles, region.Title)
	}
	return titles
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:get-cluster", &ClusterGetCmd{}))
}
