package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
)

var mapName string
var mapKey string
var mapValue string

var mapValueType string
var mapValueFile string

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "map operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	mapCmd.AddCommand(mapGetCmd)
	mapCmd.AddCommand(mapPutCmd)
	mapCmd.PersistentFlags().StringVarP(&mapName, "name", "m", "", "specify the map")
}

func getMap(clientConfig *hazelcast.Config, mapName string) (*hazelcast.Map, error) {
	ctx := context.TODO()
	var client *hazelcast.Client
	var err error
	if mapName == "" {
		return nil, errors.New("map name is required")
	}
	if clientConfig == nil {
		client, err = hazelcast.StartNewClient(ctx)
	} else {
		client, err = hazelcast.StartNewClientWithConfig(ctx, *clientConfig)
	}
	if err != nil {
		return nil, fmt.Errorf("error creating the client: %w", err)
	}
	if result, err := client.GetMap(ctx, mapName); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func retrieveFlagValues(cmd *cobra.Command) (*hazelcast.Config, error) {
	flags := cmd.InheritedFlags()
	config := internal.DefaultConfig()
	cloudToken, err := flags.GetString("cloud-token")
	if err != nil {
		return nil, err
	}
	if cloudToken != "" {
		config.Cluster.Cloud.Token = cloudToken
		config.Cluster.Cloud.Enabled = true
	} else {
		addrRaw, err := flags.GetString("address")
		if err != nil {
			return nil, err
		}
		addresses := strings.Split(addrRaw, ",")
		config.Cluster.Network.Addresses = addresses
	}
	clusterGroupName, err := flags.GetString("cluster-name")
	if err != nil {
		return nil, err
	}
	config.Cluster.Name = clusterGroupName
	return config, nil
}

func decorateCommandWithKeyFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&mapKey, "key", "k", "", "key of the map")
}

func decorateCommandWithValueFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringVarP(&mapValue, "value", "v", "", "value of the map")
	flags.StringVarP(&mapValueType, "value-type", "p", "string", "type of the value, one of: string, json")
	flags.StringVarP(&mapValueFile, "value-file", "f", "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
}
