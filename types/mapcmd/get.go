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
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const MapGetExample = `  # Get value of the given key from the map.
  hzc map get --key-type int16 --key 2012 --name myMap   # default key-type is string`

func NewGet(config *hazelcast.Config) *cobra.Command {
	var mapName, mapKey, mapKeyType string
	var showType bool
	cmd := &cobra.Command{
		Use:     "get [--name mapname | --key keyname]",
		Short:   "Get single entry from the map",
		Example: MapGetExample,
		PreRunE: hzcerrors.RequiredFlagChecker,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s, %s", mapKey, mapKeyType, err)
			}
			ci, err := getClient(cmd.Context(), config)
			if err != nil {
				return err
			}
			keyData, err := ci.EncodeData(key)
			if err != nil {
				return err
			}
			req := codec.EncodeMapGetRequest(mapName, keyData, 0)
			resp, err := ci.InvokeOnKey(cmd.Context(), req, keyData, nil)
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot get value for key %s from map %s", mapKey, mapName)
			}
			raw := codec.DecodeMapGetResponse(resp)
			valueType := raw.Type()
			value, err := ci.DecodeData(raw)
			if err != nil {
				value = serialization.NondecodedType(serialization.TypeToString(valueType))
			}
			printValueBasedOnType(cmd, value, valueType, showType)
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry")
	decorateCommandWithMapKeyTypeFlags(cmd, &mapKeyType, false)
	decorateCommandWithShowTypesFlag(cmd, &showType)
	return cmd
}
