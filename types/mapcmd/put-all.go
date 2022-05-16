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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	fds "github.com/hazelcast/hazelcast-commandline-client/types/flagdecorators"
)

func NewPutAll(config *hazelcast.Config) (*cobra.Command, error) {
	var (
		entries []types.Entry
	)
	var (
		jsonEntryPath string
		jsonEntries   map[string]interface{}
	)
	var (
		mapName string
		mapKeys,
		mapValues,
		mapValueTypes,
		mapValueFiles []string
	)
	validateJsonEntryFlag := func() error {
		if (len(mapKeys) |
			len(mapValues) |
			len(mapValueTypes) |
			len(mapValueFiles)) != 0 {
			return hzcerror.NewLoggableError(nil, fmt.Sprintf("%s is already set, there cannot be additional flags", fds.JsonEntryFlag))
		}
		return nil
	}
	validateValuesFlag := func() ([]byte, []int, error) {
		valueCount := len(mapValues) + len(mapValueFiles)
		if valueCount != len(mapKeys) {
			return nil, nil, hzcerror.NewLoggableError(nil, "number of keys and values does not match")
		}
		vOrder, tOrder := ObtainOrderingOfValueFlags(os.Args)
		if vOrder == nil {
			return nil, nil, hzcerror.NewLoggableError(nil, "correct order of values cannot be taken")
		}
		return vOrder, tOrder, nil
	}
	executePutAll := func(ctx context.Context, cmd *cobra.Command, m *hazelcast.Map, entries []types.Entry) error {
		var err error
		err = m.PutAll(ctx, entries...)
		if err != nil {
			cmd.Println("Cannot put given entries")
			isCloudCluster := config.Cluster.Cloud.Enabled
			if networkErrMsg, handled := internal.TranslateNetworkError(err, isCloudCluster); handled {
				err = hzcerror.NewLoggableError(err, networkErrMsg)
			}
			return err
		}
		return nil
	}
	cmd := &cobra.Command{
		Use:   "put-all --name mapname [--key keyname]... [[--value-file file | --value value][--value-type type]]...",
		Short: "Put values to map",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			if jsonEntryPath != "" {
				if err := validateJsonEntryFlag(); err != nil {
					return err
				}
				data, err := ioutil.ReadFile(jsonEntryPath)
				if err != nil {
					return err
				}
				if err = json.Unmarshal(data, &jsonEntries); err != nil {
					return hzcerror.NewLoggableError(err, "given json map entry file is in invalid format")
				}
				for key, jsv := range jsonEntries {
					switch jsv.(type) {
					case nil:
						// ignore null json values
					case string, float64, bool, []interface{}:
						entries = append(entries, types.Entry{Key: key, Value: jsv})
					case map[string]interface{}:
						var nJson interface{}
						mj, _ := json.Marshal(jsv)
						nJson, err = normalizeMapValue(string(mj), "", internal.TypeJSON)
						if err != nil {
							return hzcerror.NewLoggableError(err, "unknown entry value")
						}
						entries = append(entries, types.Entry{Key: key, Value: nJson})
					default:
						return hzcerror.NewLoggableError(nil, "Unknown data type in json file")
					}
				}
				m, err := getMap(ctx, config, mapName)
				if err != nil {
					return err
				}
				return executePutAll(ctx, cmd, m, entries)
			}
			vOrder, tOrder, err := validateValuesFlag()
			if err != nil {
				return err
			}
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			var (
				vC = 0
			)
			for _, key := range mapKeys {
				var t int
				curr := vOrder[0]
				var normalizedValue interface{}
				if len(tOrder) != 0 {
					t = tOrder[0]
				}
				if curr == 's' {
					v := mapValues[0]
					if len(tOrder) != 0 && t == vC {
						tv := mapValueTypes[0]
						if normalizedValue, err = normalizeMapValue(v, "", tv); err != nil {
							return err
						}
						mapValueTypes = mapValueTypes[1:]
						tOrder = tOrder[1:]
					} else {
						if normalizedValue, err = normalizeMapValue(v, "", internal.TypeString); err != nil {
							return err
						}
					}
					mapValues = mapValues[1:]
				} else {
					v := mapValueFiles[0]
					if len(tOrder) != 0 && t == vC {
						tv := mapValueTypes[0]
						if normalizedValue, err = normalizeMapValue("", v, tv); err != nil {
							return err
						}
						mapValueTypes = mapValueTypes[1:]
						tOrder = tOrder[1:]
					} else {
						if normalizedValue, err = normalizeMapValue(v, "", internal.TypeString); err != nil {
							return err
						}
					}
					mapValueFiles = mapValueFiles[1:]
				}
				entries = append(entries, types.Entry{Key: key, Value: normalizedValue})
				vOrder = vOrder[1:]
				vC++
			}
			return executePutAll(ctx, cmd, m, entries)
		},
	}
	if err := decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithMapKeyArrayFlags(cmd, &mapKeys, false, "key(s) of the map"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithMapValueArrayFlags(cmd, &mapValues, false, "value(s) of the map"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithMapValueFileArrayFlags(cmd, &mapValueFiles, false, "`path to the file that contains the value. Use \"-\" (dash) to read from stdin`"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithMapValueTypeArrayFlags(cmd, &mapValueTypes, false, "type of the value, one of: string, json"); err != nil {
		return nil, err
	}
	if err := fds.DecorateCommandWithJsonEntryFlag(cmd, &jsonEntryPath, false, "`path to json file that contains entries`"); err != nil {
		return nil, err
	}
	return cmd, nil
}
