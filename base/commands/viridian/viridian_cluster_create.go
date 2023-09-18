//go:build std || viridian

package viridian

import (
	"context"
	"errors"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterCreateCommand struct{}

func (ClusterCreateCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("create-cluster")
	long := `Creates a Viridian cluster.

Make sure you login before running this command.
`
	short := "Creates a Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagName, "", "", false, "specify the cluster name; if not given an auto-generated name is used.")
	cc.AddBoolFlag(flagDevelopment, "", false, false, "create a development cluster")
	cc.AddBoolFlag(flagPrerelease, "", false, false, "create a prerelease cluster")
	return nil
}

func (ClusterCreateCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.Metrics().Increment(metric.NewSimpleKey(), "total.viridian."+cmd.RunningMode(ec))
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	name := ec.Props().GetString(flagName)
	dev := ec.Props().GetBool(flagDevelopment)
	prerelease := ec.Props().GetBool(flagPrerelease)
	hzVersion := ec.Props().GetString(flagHazelcastVersion)
	stages := []stage.Stage[createStageState]{
		{
			ProgressMsg: "Initiating cluster creation",
			SuccessMsg:  "Initiated cluster creation",
			FailureMsg:  "Failed initiating cluster creation",
			Func: func(ctx context.Context, status stage.Statuser[createStageState]) (createStageState, error) {
				state := createStageState{}
				k8sCluster, err := getFirstAvailableK8sCluster(ctx, api)
				if err != nil {
					return state, err
				}
				cs, err := api.CreateCluster(ctx, name, getClusterType(dev), k8sCluster.ID, prerelease, hzVersion)
				if err != nil {
					return state, err
				}
				state.Cluster = cs
				return state, nil
			},
		},
		{
			ProgressMsg: "Waiting for the cluster to get ready",
			SuccessMsg:  "Cluster is ready",
			FailureMsg:  "Failed while waiting for cluster to get ready",
			Func: func(ctx context.Context, status stage.Statuser[createStageState]) (createStageState, error) {
				state := status.Value()
				if err := waitClusterState(ctx, ec, api, state.Cluster.ID, stateRunning); err != nil {
					return state, err
				}
				return state, nil
			},
		},
		{
			ProgressMsg: "Importing the configuration",
			SuccessMsg:  "Imported the configuration",
			FailureMsg:  "Failed importing the configuration",
			Func: func(ctx context.Context, status stage.Statuser[createStageState]) (createStageState, error) {
				state := status.Value()
				path, err := tryImportConfig(ctx, ec, api, state.Cluster.ID, state.Cluster.Name)
				if err != nil {
					return state, nil
				}
				state.ConfigPath = path
				return state, nil
			},
		},
	}
	state, err := stage.Execute(ctx, ec, createStageState{}, stage.NewFixedProvider(stages...))
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	ec.PrintlnUnnecessary("OK Created the cluster.\n")
	rows := output.Row{
		output.Column{
			Name:  "ID",
			Type:  serialization.TypeString,
			Value: state.Cluster.ID,
		},
	}
	if ec.Props().GetBool(clc.PropertyVerbose) {
		rows = append(rows,
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: state.Cluster.Name,
			},
			output.Column{
				Name:  "Configuration Path",
				Type:  serialization.TypeString,
				Value: state.ConfigPath,
			},
		)
	}
	return ec.AddOutputRows(ctx, rows)
}

func getClusterType(dev bool) string {
	if dev {
		return viridian.ClusterTypeDevMode
	}
	return viridian.ClusterTypeServerless
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

type createStageState struct {
	Cluster    viridian.Cluster
	ConfigPath string
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:create-cluster", &ClusterCreateCommand{}))
}
