package viridian

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
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
	cc.AddStringFlag(flagName, "", "", false, "override the cluster name")
	cc.AddStringFlag(flagPlan, "", "", false, "plan for the cluster, supported values: serverless")
	cc.AddBoolFlag(flagDevelopment, "", false, false, "start a development cluster")
	return nil
}

func (cm ClusterCreateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	name := ec.Props().GetString(flagName)
	plan := ec.Props().GetString(flagPlan)
	if err := validatePlan(plan); err != nil {
		return err
	}
	dev := ec.Props().GetBool(flagDevelopment)
	csi, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Retrieving available Kubernetes clusters")
		K8sCluster, err := getFirstAvailableK8sCluster(ctx, api)
		if err != nil {
			return nil, err
		}
		sp.SetText("Creating the cluster")
		cs, err := api.CreateCluster(ctx, name, viridian.ClusterPlan(strings.ToUpper(plan)), K8sCluster.ID, dev)
		if err != nil {
			return nil, err
		}
		return cs, nil
	})
	if err != nil {
		ec.Logger().Error(err)
		return fmt.Errorf("error creating a cluster. Did you login?: %w", err)
	}
	stop()
	cs := csi.(viridian.Cluster)
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	if verbose {
		row := output.Row{
			output.Column{
				Name:  "ID",
				Type:  serialization.TypeString,
				Value: cs.ID,
			},
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: cs.Name,
			},
		}
		return ec.AddOutputRows(ctx, row)
	}
	return nil
}

func validatePlan(plan string) error {
	switch plan {
	case "", "serverless":
		return nil
	default:
		return errors.New("plan invalid: possible values: [serverless]")
	}
}
func getFirstAvailableK8sCluster(ctx context.Context, api *viridian.API) (viridian.K8sCluster, error) {
	clusters, err := api.ListAvailableK8sClusters(ctx)
	if err != nil {
		return viridian.K8sCluster{}, err
	}
	if len(clusters) == 0 {
		return viridian.K8sCluster{}, errors.New("there is no available K8s cluster")
	}
	return clusters[0], nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:cluster-create", &ClusterCreateCmd{}))
}
