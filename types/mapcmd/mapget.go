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
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const MapGetExample = `map get --key hello --name myMap
map get --key-type int16 --key 2012 --name yearbook`

func NewGet(config *hazelcast.Config) *cobra.Command {
	var mapName, mapKey, mapKeyType string
	cmd := &cobra.Command{
		Use:     "get [--name mapname | --key keyname]",
		Short:   "Get from map",
		Example: MapGetExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s", mapKey, mapKeyType)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			value, err := m.Get(ctx, key)
			if err != nil {
				isCloudCluster := config.Cluster.Cloud.Enabled
				if networkErrMsg, handled := hzcerrors.TranslateNetworkError(err, isCloudCluster); handled {
					return hzcerrors.NewLoggableError(err, networkErrMsg)
				}
				return hzcerrors.NewLoggableError(err, "Cannot get value for key %s from map %s", mapKey, mapName)
			}
			if value == nil {
				cmd.Println("There is no value corresponding to the provided key")
				return nil
			}
			switch v := value.(type) {
			case serialization.JSON:
				if err := quick.Highlight(cmd.OutOrStdout(), fmt.Sprintln(v.String()),
					"json", "terminal", "tango"); err != nil {
					cmd.Println(v.String())
				}
			default:
				cmd.Println(value)
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName)
	decorateCommandWithKeyFlags(cmd, &mapKey, &mapKeyType)
	return cmd
}
