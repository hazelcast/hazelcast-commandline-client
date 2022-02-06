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
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "map {get | put} --name mapname --key keyname [--value-type type | --value-file file | --value value]",
		Short: "map operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewGet(), NewPut())
	return cmd
}

func getMap(ctx context.Context, clientConfig *hazelcast.Config, mapName string) (result *hazelcast.Map, err error) {
	hzcClient, err := internal.ConnectToCluster(ctx, clientConfig)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error creating the client: %w", err)
	}
	if result, err = hzcClient.GetMap(ctx, mapName); err != nil {
		errorMsg := err.Error()
		if msg, isHandled := internal.TranslateNetworkError(err, clientConfig.Cluster.Cloud.Enabled); isHandled {
			errorMsg = msg
		}
		fmt.Println("Error:", errorMsg)
		return nil, err
	}
	return
}

func decorateCommandWithMapNameFlags(cmd *cobra.Command, mapName *string) {
	cmd.Flags().StringVarP(mapName, "name", "n", "", "specify the map name")
	cmd.MarkFlagRequired("name")
}

func decorateCommandWithKeyFlags(cmd *cobra.Command, mapKey *string) {
	cmd.Flags().StringVarP(mapKey, "key", "k", "", "key of the map")
	cmd.MarkFlagRequired("key")
	cmd.RemoveCommand()
}

func decorateCommandWithValueFlags(cmd *cobra.Command, mapValue, mapValueFile, mapValueType *string) {
	flags := cmd.Flags()
	flags.StringVarP(mapValue, "value", "v", "", "value of the map")
	flags.StringVarP(mapValueFile, "value-file", "f", "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
	flags.StringVarP(mapValueType, "value-type", "t", "string", "type of the value, one of: string, json")
	cmd.RegisterFlagCompletionFunc("value-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "string"}, cobra.ShellCompDirectiveDefault
	})
}
