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
)

const MapEntriesExample = `  # Get all entries from the map with given delimiter (default tab character).
  hzc map entries -n mapname --delim ":"`

func NewEntries(config *hazelcast.Config) *cobra.Command {
	var (
		delim,
		mapName string
	)
	cmd := &cobra.Command{
		Use:     "entries --name mapname [--delim delimiter]]",
		Short:   "Get all entries from the map with given delimiter",
		Example: MapEntriesExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var entries []types.Entry
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			entries, err = m.GetEntrySet(cmd.Context())
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot get entries for the given keys for map %s", mapName)
			}
			for _, entry := range entries {
				cmd.Printf("%s%s%s\n", formatGoTypeToOutput(entry.Key), delim, formatGoTypeToOutput(entry.Value))
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithDelimiter(cmd, &delim, false, "delimiter of printed key, value pairs")
	return cmd
}
