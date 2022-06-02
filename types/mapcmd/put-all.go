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
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	fds "github.com/hazelcast/hazelcast-commandline-client/internal/flagdecorators"
)

const MapPutAllExample = `  # Put key, value pairs to map.
  hzc map put-all -n mapname -k k1 -v v1 -k k2 -v v2
  
  # Put key, value pairs to map in another order.
  hzc map put-all -n mapname -k k1 -k k2 -v v1 -v v2
  
  # Put key, value pairs to map but one of the value type is json file.
  hzc map put-all -n mapname -k k1 -f valueFile.json -t json -k k2 -v v2

  # Put all key, value pairs to map from the entry json file .
  hzc map put-all -n mapname --json-entry entries.json

  # Example json entry file
  {
    "key1": "value1",
    "key2": {
      "innerData": "data",
      "anotherInnerData": 5.0
    },
    "key3": true,
    "key4": [1, 2, 3, 4, 5]
  }
  - Entries with "null" values are being ignored.
  
  # Coupling rule of keys and values given in different order
  - Keys and values are coupled according to the order they are provided. That means the first given key will be matched with the first given 
  value from left to right. Therefore, keys and values (given through file or directly from the command line) must be equal in number.
  - BUT, for "--type" flag, this rule is not applied. Type of the value flag is given just after the actual value flag.
`

func NewPutAll(config *hazelcast.Config) *cobra.Command {
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
		if len(mapKeys) != 0 ||
			len(mapValues) != 0 ||
			len(mapValueTypes) != 0 ||
			len(mapValueFiles) != 0 {
			return hzcerrors.NewLoggableError(nil, fmt.Sprintf("%s is already set, there cannot be additional flags", fds.JsonEntryFlag))
		}
		return nil
	}
	validateValuesFlag := func() ([]byte, []int, error) {
		valueCount := len(mapValues) + len(mapValueFiles)
		if valueCount != len(mapKeys) {
			return nil, nil, hzcerrors.NewLoggableError(nil, "number of keys and values does not match")
		}
		vOrder, tOrder := ObtainOrderingOfValueFlags(os.Args)
		if vOrder == nil {
			return nil, nil, hzcerrors.NewLoggableError(nil, "correct order of values cannot be taken")
		}
		return vOrder, tOrder, nil
	}
	executePutAll := func(ctx context.Context, cmd *cobra.Command, m *hazelcast.Map, entries []types.Entry) error {
		var err error
		err = m.PutAll(ctx, entries...)
		if err != nil {
			cmd.Println("Cannot put given entries")
			isCloudCluster := config.Cluster.Cloud.Enabled
			if networkErrMsg, handled := hzcerrors.TranslateNetworkError(err, isCloudCluster); handled {
				err = hzcerrors.NewLoggableError(err, networkErrMsg)
			}
			return err
		}
		return nil
	}
	cmd := &cobra.Command{
		Use:        "put-all [--name mapname | {[[--key keyname]... | [[--value-file file | --value value][--value-type type]]...] | [--json-entry jsonEntryFile]}]",
		Aliases:    nil,
		SuggestFor: nil,
		Short:      "Put values to map",
		Long:       "",
		Example:    MapPutAllExample,
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
					return hzcerrors.NewLoggableError(err, "given json map entry file is in invalid format")
				}
				for key, jsv := range jsonEntries {
					switch jsv.(type) {
					case nil:
						// ignore null json values
					case []interface{}:
						result, _ := jsv.([]interface{})
						for i, item := range result {
							if e, ok := item.(map[string]interface{}); ok {
								mj, _ := json.Marshal(e)
								nJson := serialization.JSON(mj)
								result[i] = nJson
							}
						}
						entries = append(entries, types.Entry{Key: key, Value: result})
					case string, float64, bool:
						entries = append(entries, types.Entry{Key: key, Value: jsv})
					case map[string]interface{}:
						var nJson interface{}
						mj, _ := json.Marshal(jsv)
						nJson = serialization.JSON(mj)
						entries = append(entries, types.Entry{Key: key, Value: nJson})
					default:
						return hzcerrors.NewLoggableError(nil, "Unknown data type in json file")
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
						if normalizedValue, err = normalizeMapValue(v, "", internal.TypeNameString); err != nil {
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
						if normalizedValue, err = normalizeMapValue(v, "", internal.TypeNameString); err != nil {
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
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyArrayFlags(cmd, &mapKeys, false, "key(s) of the map")
	decorateCommandWithMapValueArrayFlags(cmd, &mapValues, false, "value(s) of the map")
	decorateCommandWithMapValueFileArrayFlags(cmd, &mapValueFiles, false,
		"`path to the file that contains the value. Use \"-\" (dash) to read from stdin`")
	decorateCommandWithMapValueTypeArrayFlags(cmd, &mapValueTypes, false, "type of the value, one of: string, json")
	fds.DecorateCommandWithJsonEntryFlag(cmd, &jsonEntryPath, false, "`path to json file that contains entries`")
	return cmd
}
