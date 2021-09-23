/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"log"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"
)

var mapName string
var mapKey string
var mapValue string

var mapValueType string
var mapValueFile string

var MapCmd = &cobra.Command{
	Use:   "map {get | put} --name mapname --key keyname [--value-type type | --value-file file | --value value]",
	Short: "map operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	MapCmd.AddCommand(mapGetCmd)
	MapCmd.AddCommand(mapPutCmd)
}

func getMap(clientConfig *hazelcast.Config, mapName string) (*hazelcast.Map, error) {
	var client *hazelcast.Client
	var err error
	defer func() {
		obj := recover()
		if err, ok := obj.(error); ok {
			networkErr := internal.HandleNetworkError(err)
			log.Fatal(networkErr)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if mapName == "" {
		return nil, errors.New("map name is required")
	}
	if clientConfig == nil {
		clientConfig = &hazelcast.Config{}
	}
	clientConfig.Cluster.ConnectionStrategy.Retry.InitialBackoff = types.Duration(1 * time.Second)
	client, err = hazelcast.StartNewClientWithConfig(ctx, *clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating the client: %w", err)
	}
	if result, err := client.GetMap(ctx, mapName); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func decorateCommandWithMapNameFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&mapName, "name", "m", "", "specify the map name")
	cmd.MarkFlagRequired("name")
}

func decorateCommandWithKeyFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&mapKey, "key", "k", "", "key of the map")
	cmd.MarkFlagRequired("key")
}

func decorateCommandWithValueFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&mapValue, "value", "v", "", "value of the map")
	flags.StringVarP(&mapValueType, "value-type", "t", "string", "type of the value, one of: string, json")
	flags.StringVarP(&mapValueFile, "value-file", "f", "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
	cmd.RegisterFlagCompletionFunc("value-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "string"}, cobra.ShellCompDirectiveDefault
	})
}
