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
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	fds "github.com/hazelcast/hazelcast-commandline-client/types/flagdecorators"
)

func NewGetAll(config *hazelcast.Config) (*cobra.Command, error) {
	var (
		delim string
	)
	var (
		mapName string
		mapKeys []string
	)
	validateFlags := func() error {
		if len(mapKeys) == 0 {
			return hzcerror.NewLoggableError(nil, "At least one key must be given")
		}
		return nil
	}
	cmd := &cobra.Command{
		Use:   "get-all [--name mapname | [--key keyname]... [--delim delimiter]]",
		Short: "Get all matched entries from specified map",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			var err error
			if err = validateFlags(); err != nil {
				return err
			}
			var m *hazelcast.Map
			m, err = getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			keys := make([]interface{}, len(mapKeys))
			for i := range mapKeys {
				keys[i] = mapKeys[i]
			}
			var entries []types.Entry
			entries, err = m.GetAll(ctx, keys...)
			if err != nil {
				var handled bool
				handled, err = cloudcb(err, config)
				if handled {
					return err
				}
				return hzcerror.NewLoggableError(err, "Cannot get entries for the given keys for map %s", mapName)
			}
			for _, entry := range entries {
				cmd.Print(entry.Key, delim)
				switch v := entry.Value.(type) {
				case serialization.JSON:
					if err = quick.Highlight(cmd.OutOrStdout(), fmt.Sprintln(v.String()),
						"json", "terminal", "tango"); err != nil {
						cmd.Println(v.String())
					}
				default:
					cmd.Println(v)
				}
			}
			return nil
		},
	}
	if err := decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithMapKeyArrayFlags(cmd, &mapKeys, false, "key(s) of the entry"); err != nil {
		return nil, err
	}
	if err := fds.DecorateCommandWithDelimiter(cmd, &delim, false, "delimiter of printed key, value pairs"); err != nil {
		return nil, err
	}
	return cmd, nil
}
