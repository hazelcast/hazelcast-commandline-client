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
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

const MapKeysExample = `  # Get all the keys from the map.
  hzc keys -n mapname`

func NewKeys(config *hazelcast.Config) *cobra.Command {
	var mapName string
	cmd := &cobra.Command{
		Use:     "keys --name mapname",
		Short:   "Get all the keys from the map",
		Example: MapKeysExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var keys []interface{}
			var m *hazelcast.Map
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			keys, err = m.GetKeySet(cmd.Context())
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot get entries for the given keys for map %s", mapName)
			}
			for _, k := range keys {
				cmd.Println(formatGoTypeToOutput(k))
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	return cmd
}
