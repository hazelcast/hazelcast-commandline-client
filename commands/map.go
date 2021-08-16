package commands

import (
	"context"
	"errors"
	"fmt"

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
	mapCmd.PersistentFlags().StringVar(&mapName, "name", "", "specify the map")
	rootCmd.AddCommand(mapCmd)
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

func decorateCommandWithKeyFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&mapKey, "key", "", "key of the map")
}

func decorateCommandWithValueFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringVar(&mapValue, "value", "", "value of the map")
	flags.StringVar(&mapValueType, "value-type", "string", "type of the value, one of: string, json")
	flags.StringVar(&mapValueFile, "value-file", "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
}
