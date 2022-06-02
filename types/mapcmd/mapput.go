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
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const MapPutExample = `map put --key hello --value world --name myMap    #puts entry into map directly
map put --key-type string --key hello --value-type float32 --value 19.94 --name myMap`

func NewPut(config *hazelcast.Config) *cobra.Command {
	// flags
	var (
		mapName,
		mapKey,
		mapKeyType,
		mapValue,
		mapValueType,
		mapValueFile string
	)
	cmd := &cobra.Command{
		Use:     "put [--name mapname | --key keyname | --value-type type | --value-file file | --value value]",
		Short:   "Put value to map",
		Example: MapPutExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s", mapKey, mapKeyType)
			}
			var normalizedValue interface{}
			if normalizedValue, err = normalizeMapValue(mapValue, mapValueFile, mapValueType); err != nil {
				return err
			}
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			_, err = m.Put(ctx, key, normalizedValue)
			if err == nil {
				return err
			}
			cmd.Printf("Cannot put value for key %s to map %s\n", mapKey, mapName)
			isCloudCluster := config.Cluster.Cloud.Enabled
			if networkErrMsg, handled := hzcerrors.TranslateNetworkError(err, isCloudCluster); handled {
				err = hzcerrors.NewLoggableError(err, networkErrMsg)
			}
			return err
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName)
	decorateCommandWithKeyFlags(cmd, &mapKey, &mapKeyType)
	decorateCommandWithValueFlags(cmd, &mapValue, &mapValueFile, &mapValueType)
	return cmd
}

func normalizeMapValue(v, vFile, vType string) (interface{}, error) {
	var valueStr string
	var err error
	switch {
	case v != "" && vFile != "":
		return nil, hzcerrors.NewLoggableError(nil, "Only one of --value and --value-file must be specified")
	case v != "":
		valueStr = v
	case vFile != "":
		if valueStr, err = loadValueFile(vFile); err != nil {
			err = hzcerrors.NewLoggableError(err, "Cannot load the value file. Make sure file exists and process has correct access rights")
		}
	default:
		err = hzcerrors.NewLoggableError(nil, "One of the value flags (--value or --value-file) must be set")
	}
	if err != nil {
		return nil, err
	}
	mapValue, err := internal.ConvertString(valueStr, vType)
	if err != nil {
		err = hzcerrors.NewLoggableError(err, "Conversion error on value %s to value-type %s", valueStr, vType)
	}
	return mapValue, err
}

func loadValueFile(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}
	if path == "-" {
		if value, err := ioutil.ReadAll(os.Stdin); err != nil {
			return "", err
		} else {
			return string(value), nil
		}
	}
	if value, err := ioutil.ReadFile(path); err != nil {
		return "", err
	} else {
		return string(value), nil
	}
}
