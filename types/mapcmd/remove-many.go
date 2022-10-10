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
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func NewRemoveMany(config *hazelcast.Config) *cobra.Command {
	var (
		mapName, mapKeyType string
		mapKeys             []string
	)
	cmd := &cobra.Command{
		Use:   "remove-many --name mapname [--key-type keysType] --key keyname [--key keyname2...]",
		Short: "Removes entries from the map corresponding to the given keys",
		Example: `  # Remove entries from the map
  hzc map remove-many -n mapname -k k1 -k k2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// actual keys
			var keys []interface{}
			for _, mk := range mapKeys {
				key, err := internal.ConvertString(mk, mapKeyType)
				if err != nil {
					return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s, %s", mk, mapKeyType, err)
				}
				keys = append(keys, key)
			}
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			var errs []string
			for i, k := range keys {
				_, err = m.Remove(cmd.Context(), k)
				if err != nil {
					var handled bool
					handled, err = isCloudIssue(err, config)
					if !handled {
						vErr := hzcerrors.NewLoggableError(err, "Cannot remove key %s from map %s", mapKeys[i], mapName)
						errs = append(errs, vErr.VerboseError())
					}
					errs = append(errs, err.Error())
				}
			}
			if len(errs) != 0 {
				return hzcerrors.NewLoggableError(nil, "Following errors encountered:\n%s", strings.Join(errs, "\n"))
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyArrayFlags(cmd, &mapKeys, true, "keys of the entries to remove")
	decorateCommandWithMapKeyTypeFlags(cmd, &mapKeyType, false)
	return cmd
}
