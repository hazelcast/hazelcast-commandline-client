package commands

import (
	"errors"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client/v4/hazelcast"
	"github.com/spf13/cobra"
)

var mapName string
var mapKey string
var mapValue string

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Map operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	mapCmd.AddCommand(mapGetCmd)
	mapCmd.AddCommand(mapPutCmd)
	mapCmd.PersistentFlags().StringVar(&mapName, "name", "", "specify the map")
}

func getMap(clientConfig *hazelcast.Config, mapName string) (hazelcast.Map, error) {
	var client hazelcast.Client
	var err error
	if mapName == "" {
		return nil, errors.New("map name is required")
	}
	if clientConfig == nil {
		client, err = hazelcast.NewClient()
	} else {
		client, err = hazelcast.NewClientWithConfig(clientConfig)
	}
	if err != nil {
		return nil, fmt.Errorf("error creating the client: %w", err)
	}
	if result, err := client.GetMap(mapName); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}
