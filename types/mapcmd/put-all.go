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

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const MapPutAllExample = `  # Put key, value pairs while specifying types of both keys and values
  # Keys and values are matched with the given order
  hzc map put-all -n mapname --key-type int16 -k 1 -k 2 --value-type json -f valueFile.json -v '{"field":"tmp"}' 

  # Put all key, value pairs to map from the entry json file. Null values are ignored.
  hzc map put-all -n mapname --json-entry entries.json
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
		mapKeyType,
		mapValueType,
		mapName string
		mapKeys,
		mapValues,
		mapValueFiles []string
	)
	validateJsonEntryFlag := func() error {
		if len(mapKeys) != 0 ||
			len(mapValues) != 0 ||
			len(mapValueFiles) != 0 ||
			mapKeyType != "" ||
			mapValueType != "" {
			return hzcerrors.NewLoggableError(nil, fmt.Sprintf("%s is already set, there cannot be additional flags", internal.JSONEntryFlag))
		}
		return nil
	}
	validateValuesFlag := func() ([]byte, error) {
		vOrder := ObtainOrderingOfValueFlags(os.Args)
		if len(vOrder) == 0 {
			return nil, hzcerrors.NewLoggableError(nil, "correct order of values cannot be taken")
		}
		return vOrder, nil
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
			var err error
			if jsonEntryPath != "" {
				if err := validateJsonEntryFlag(); err != nil {
					return err
				}
				// assign default values
				if mapKeyType == "" {
					mapKeyType = "string"
				}
				if mapValueType == "" {
					mapKeyType = "string"
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
				m, err := getMap(cmd.Context(), config, mapName)
				if err != nil {
					return err
				}
				return executePutAll(cmd.Context(), cmd, m, entries)
			}
			valueNumber := len(mapValues) + len(mapValueFiles)
			if valueNumber != len(mapKeys) {
				return hzcerrors.NewLoggableError(nil, "number of keys and values do not match")
			}
			var vOrder []byte
			vOrder, err = validateValuesFlag()
			if err != nil {
				return err
			}
			for _, key := range mapKeys {
				curr := vOrder[0]
				var normalizedValue interface{}
				if curr == 's' {
					v := mapValues[0]
					if normalizedValue, err = normalizeMapValue(v, "", mapValueType); err != nil {
						return err
					}
					mapValues = mapValues[1:]
				} else {
					v := mapValueFiles[0]
					if normalizedValue, err = normalizeMapValue("", v, mapValueType); err != nil {
						return err
					}
					mapValueFiles = mapValueFiles[1:]
				}
				vOrder = vOrder[1:]
				var normalizedKey interface{}
				if mapKeyType != "" {
					normalizedKey, err = internal.ConvertString(key, mapKeyType)
					if err != nil {
						return hzcerrors.NewLoggableError(err, "key type does cannot be converted")
					}
				} else {
					normalizedKey = key
				}
				entries = append(entries, types.Entry{Key: normalizedKey, Value: normalizedValue})
			}
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			return executePutAll(cmd.Context(), cmd, m, entries)
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyArrayFlags(cmd, &mapKeys, false, "key(s) of the map")
	decorateCommandWithMapKeyTypeFlags(cmd, &mapKeyType, false)
	decorateCommandWithMapValueArrayFlags(cmd, &mapValues, false, "value(s) of the map")
	decorateCommandWithMapValueFileArrayFlags(cmd, &mapValueFiles, false,
		`path to the file that contains the value. Use "-" (dash) to read from stdin`)
	decorateCommandWithMapValueTypeFlags(cmd, &mapValueType, false)
	internal.DecorateCommandWithJSONEntryFlag(cmd, &jsonEntryPath, false, `path to json file that contains entries`)
	return cmd
}
