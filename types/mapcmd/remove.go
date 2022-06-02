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
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
)

func NewRemove(config *hazelcast.Config) *cobra.Command {
	var (
		mapName,
		mapKey string
	)
	cmd := &cobra.Command{
		Use:   "remove [--name mapname | --key keyname]",
		Short: "Remove key(s)",
		Example: `  # Remove key from the map if it exists.
  hzc map remove -n mapname -k k1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// same context timeout for both single entry removal and map cleanup
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			var err error
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			_, err = m.Remove(ctx, mapKey)
			if err != nil {
				var handled bool
				handled, err = cloudcb(err, config)
				if handled {
					return err
				}
				return hzcerror.NewLoggableError(err, "Cannot remove given key from map %s", mapName)
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry")
	return cmd
}
