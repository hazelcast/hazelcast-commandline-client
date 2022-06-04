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
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const MapGetAllExample = `  # Get matched entries from the map with default delimiter. Default delimiter is the tab character.
  hzc get-all -n mapname -k k1 -k k2

  # Get matched entries from the map with custom delimiter.
  hzc get-all -n mapname -k k1 -k k2 --delim ":"`

func NewGetAll(config *hazelcast.Config) *cobra.Command {
	var (
		delim,
		mapName string
		mapKeyTypes,
		mapKeys []string
	)
	validateFlags := func() error {
		if len(mapKeys) == 0 {
			return hzcerrors.NewLoggableError(nil, "At least one key must be given")
		}
		return nil
	}
	cmd := &cobra.Command{
		Use:     "get-all [--name mapname | [--key keyname]... [--delim delimiter]]",
		Short:   "Get all matched entries from the map",
		Example: MapGetAllExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if err = validateFlags(); err != nil {
				return err
			}
			keys := make([]interface{}, len(mapKeys))
			for i := range mapKeys {
				if len(mapKeyTypes) != 0 {
					keys[i], err = internal.ConvertString(mapKeys[i], mapKeyTypes[0])
					if err != nil {
						return hzcerrors.NewLoggableError(err, "key type does cannot be converted")
					}
					mapKeyTypes = mapKeyTypes[1:]
				} else {
					keys[i] = mapKeys[i]
				}
			}
			var entries []types.Entry
			var m *hazelcast.Map
			m, err = getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			entries, err = m.GetAll(cmd.Context(), keys...)
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot get entries for the given keys for map %s", mapName)
			}
			for _, entry := range entries {
				cmd.Print(entry.Key, delim)
				printValueBasedOnType(cmd, entry.Value)
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyArrayFlags(cmd, &mapKeys, true, "key(s) of the entry")
	decorateCommandWithMapKeyTypeArrayFlags(cmd, &mapKeyTypes, false, "type of the key")
	decorateCommandWithDelimiter(cmd, &delim, false, "delimiter of printed key, value pairs")
	return cmd
}
