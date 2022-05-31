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
package mapcmd

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func New(config *hazelcast.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "map {get | put} --name mapname --key keyname [--value-type type | --value-file file | --value value]",
		Short:   "Map operations",
		Example: fmt.Sprintf("%s\n%s\n%s", MapPutExample, MapGetExample, MapUseExample),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// All the following lines are to set map name if it is set by "use" command.
			// If the map name is given explicitly, do not set the one given with "use" command.
			// Missing flag errors are not handled here.
			// They are expected to be handled by the actual command.
			persister := internal.PersistedNamesFromContext(cmd.Context())
			val, isSet := persister["map"]
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
				return hzcerrors.NewLoggableError(err, "Default name for map cannot be set")
			}
			return nil
		},
	}
	cmd.AddCommand(NewGet(config), NewPut(config), NewUse())
	return cmd
}

func getMap(ctx context.Context, clientConfig *hazelcast.Config, mapName string) (result *hazelcast.Map, err error) {
	hzcClient, err := internal.ConnectToCluster(ctx, clientConfig)
	if err != nil {
		return nil, hzcerrors.NewLoggableError(err, "Cannot get initialize client")
	}
	if result, err = hzcClient.GetMap(ctx, mapName); err != nil {
		if msg, isHandled := hzcerrors.TranslateNetworkError(err, clientConfig.Cluster.Cloud.Enabled); isHandled {
			err = hzcerrors.NewLoggableError(err, msg)
		}
		return nil, err
	}
	return
}

const MapUseExample = "map use m1    # sets the default map name to m1 unless set explicitly"

func NewUse() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `use [map-name | --reset]`,
		Short: "sets default map name",
		Example: MapUseExample + `
map get --key k1    # --name m1\" is inferred
map use --reset	    # resets the behaviour`,
		RunE: func(cmd *cobra.Command, args []string) error {
			persister := internal.PersistedNamesFromContext(cmd.Context())
			if cmd.Flags().Changed("reset") {
				delete(persister, "map")
				return nil
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			if len(args) > 1 {
				cmd.Println("Provide map name between \"\" quotes if it contains white space")
				return nil
			}
			persister["map"] = args[0]
			return nil
		},
	}
	_ = cmd.Flags().BoolP("reset", "", false, "unset default name for map")
	return cmd
}

func decorateCommandWithMapNameFlags(cmd *cobra.Command, mapName *string) {
	cmd.Flags().StringVarP(mapName, "name", "n", "", "specify the map name")
	cmd.MarkFlagRequired("name")
}

func decorateCommandWithKeyFlags(cmd *cobra.Command, mapKey, mapKeyType *string) {
	flags := cmd.Flags()
	flags.StringVarP(mapKey, "key", "k", "", "key of the map")
	flags.StringVarP(mapKeyType, "key-type", "", "string", "type of the key, one of: string, json, int(8,16,32,64), float(32,64)")
	cmd.MarkFlagRequired("key")
	cmd.RegisterFlagCompletionFunc("key-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return internal.SupportedTypes, cobra.ShellCompDirectiveDefault
	})
}

func decorateCommandWithValueFlags(cmd *cobra.Command, mapValue, mapValueFile, mapValueType *string) {
	flags := cmd.Flags()
	flags.StringVarP(mapValue, "value", "v", "", "value of the map")
	flags.StringVarP(mapValueFile, "value-file", "f", "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
	flags.StringVarP(mapValueType, "value-type", "t", "string", "type of the value, one of: string, json, int(8,16,32,64), float(32,64)")
	cmd.RegisterFlagCompletionFunc("value-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return internal.SupportedTypes, cobra.ShellCompDirectiveDefault
	})
}
