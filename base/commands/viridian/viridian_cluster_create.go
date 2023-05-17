package viridian

import (
	"context"
	"errors"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterCreateCmd struct{}

func (cm ClusterCreateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("create-cluster [flags]")
	long := `Creates a Viridian cluster.

Make sure you login before running this command.
`
	short := "Creates a Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 0)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagName, "", "", false, "specify the cluster name; if not given an auto-generated name is used.")
	cc.AddStringFlag(flagClusterType, "", "", false, "type for the cluster")
	cc.AddStringFlag(flagHazelcastVersion, "", "", false, "version of the Hazelcast cluster")
	return nil
}

func (cm ClusterCreateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	name := ec.Props().GetString(flagName)
	clusterType := ec.Props().GetString(flagClusterType)
	hzVersion := ec.Props().GetString(flagHazelcastVersion)
	csi, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Creating the cluster")
		k8sCluster, err := getFirstAvailableK8sCluster(ctx, api)
		if err != nil {
			return nil, err
		}
		cs, err := api.CreateCluster(ctx, name, clusterType, k8sCluster.ID, hzVersion)
		if err != nil {
			return nil, err
		}
		return cs, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	c := csi.(viridian.Cluster)
	tryImportConfig(ctx, ec, api, c)
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	if verbose {
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
		}
		return ec.AddOutputRows(ctx, row)
	}
	return nil
}

func getFirstAvailableK8sCluster(ctx context.Context, api *viridian.API) (viridian.K8sCluster, error) {
	clusters, err := api.ListAvailableK8sClusters(ctx)
	if err != nil {
		return viridian.K8sCluster{}, err
	}
	if len(clusters) == 0 {
		return viridian.K8sCluster{}, errors.New("cluster creation is not available, try again later")
	}
	return clusters[0], nil
}

func tryImportConfig(ctx context.Context, ec plug.ExecContext, api *viridian.API, cluster viridian.Cluster) {
	cfgName := cluster.Name
	cp, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Waiting for the cluster to get ready")
		if err := waitClusterState(ctx, ec, api, cluster.ID, stateRunning); err != nil {
			// do not import the config and exit early
			return nil, err
		}
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
		ec.Logger().Error(err)
		return
	}
	stop()
	ec.Logger().Info("Imported configuration %s and saved to: %s", cfgName, cp)
	ec.PrintlnUnnecessary(fmt.Sprintf("Imported configuration: %s", cfgName))
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:create-cluster", &ClusterCreateCmd{}))
}
