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

	"github.com/hazelcast/hazelcast-commandline-client/commands/common"
	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "map {get | put} --name mapname --key keyname [--value-type type | --value-file file | --value value]",
		Short:   "map operations",
		Example: fmt.Sprintf("%s\n%s", MapPutExample, MapGetExample),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			persister := common.PersisterFromContext(cmd.Context())
			val, isSet := persister.Get("map")
			if !isSet {
				return nil
			}
			nameFlag := cmd.Flag("name")
			if nameFlag == nil {
				// flag is absent
				return nil
			}
			if nameFlag.Changed {
				// flag value is set explicitly
				return nil
			}
			if err := cmd.Flags().Set("name", val); err != nil {
				return hzcerror.NewLoggableError(err, "Default name for map cannot be set")
			}
			return nil
		},
	}
	cmd.AddCommand(NewGet(), NewPut(), NewUse())
	return cmd
}

func NewUse() *cobra.Command {
	//var mapName, mapKey string
	cmd := &cobra.Command{
		Use:   "use map-name",
		Short: "set default map name",
		RunE: func(cmd *cobra.Command, args []string) error {
			persister := common.PersisterFromContext(cmd.Context())
			if cmd.Flags().Changed("reset") {
				persister.Reset("map")
				return nil
			}
			if len(args) == 0 {
				cmd.Println("Default map name is not provided")
				return nil
			}
			if len(args) > 1 {
				cmd.Println("Provide map name between \"\" quotes if it contains white space")
				return nil
			}
			persister.Set("map", args[0])
			return nil
		},
	}
	_ = cmd.Flags().BoolP("reset", "", false, "unset default name for map")
	return cmd
}

func getMap(ctx context.Context, clientConfig *hazelcast.Config, mapName string) (result *hazelcast.Map, err error) {
	hzcClient, err := internal.ConnectToCluster(ctx, clientConfig)
	if err != nil {
		return nil, err
	}
	if result, err = hzcClient.GetMap(ctx, mapName); err != nil {
		if msg, isHandled := internal.TranslateNetworkError(err, clientConfig.Cluster.Cloud.Enabled); isHandled {
			err = hzcerror.NewLoggableError(err, msg)
		}
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
