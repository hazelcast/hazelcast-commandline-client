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

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/types/flagdecorators"
)

func NewGet(config *hazelcast.Config) *cobra.Command {
	var all bool
	var (
		resultOutput MapRequestedOutput
		output       string
	)
	var (
		mapName string
		mapKeys []string
	)
	validateFlags := func() error {
		if len(mapKeys) == 0 && !all {
			return hzcerror.NewLoggableError(nil, "At least one key must be given")
		}
		if all && len(mapKeys) != 0 {
			return hzcerror.NewLoggableError(nil, "all flag and key flag cannot be in the same command")
		}
		switch output {
		case MapOutputKeys.String():
			resultOutput = MapOutputKeys
		case MapOutputValues.String():
			resultOutput = MapOutputValues
		case MapOutputEntries.String():
			resultOutput = MapOutputEntries
		default:
			return hzcerror.NewLoggableError(nil, "Undefined output option")
		}
		return nil
	}
	cmd := &cobra.Command{
		Use:   "get [--name mapname | --key keyname]",
		Short: "Get from map",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			if err := validateFlags(); err != nil {
				return err
			}
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			switch resultOutput {
			case MapOutputKeys:
				var keys []interface{}
				keys, err = m.GetKeySet(ctx)
				if err != nil {
					return err
				}
				if all {
					for _, key := range keys {
						fmt.Printf("> Key: %v\n", key)
					}
				} else {
					for _, key := range keys {
						if contains(mapKeys, key) {
							fmt.Printf("> Key: %v\n", key)
						}
					}
				}
			case MapOutputValues:
				if all {
					var values []interface{}
					values, err = m.GetValues(ctx)
					if err != nil {
						return err
					}
					for _, value := range values {
						fmt.Printf("> Value: %v\n", value)
					}
				} else {
					var entries []types.Entry
					entries, err = m.GetEntrySet(ctx)
					if err != nil {
						return err
					}
					for _, entry := range entries {
						if contains(mapKeys, entry.Key) {
							fmt.Printf("> Value: %v\n", entry.Value)
						}
					}
				}
			case MapOutputEntries:
				var entries []types.Entry
				entries, err = m.GetEntrySet(ctx)
				if err != nil {
					return err
				}
				if all {
					for _, entry := range entries {
						fmt.Printf("> Key: %v\n  Value: %v\n", entry.Key, entry.Value)
					}
				} else {
					for _, entry := range entries {
						if contains(mapKeys, entry.Key) {
							fmt.Printf("> Key: %v\n  Value: %v\n", entry.Key, entry.Value)
						}
					}
				}
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName)
	decorateCommandWithMapKeySliceFlags(cmd, &mapKeys, false, "key(s) of the map")
	decorateCommandWithMapOutputFlag(cmd, &output, false,
		fmt.Sprintf("Output options. It can be %s, %s or %s", MapOutputKeys, MapOutputValues, MapOutputEntries))
	flagdecorators.DecorateCommandWithAllFlag(cmd, &all, false, "represent all entry in given context")
	return cmd
}

func contains(s []string, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
