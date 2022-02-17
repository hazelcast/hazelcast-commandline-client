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
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func NewPut() *cobra.Command {
	// flags
	var (
		mapName,
		mapKey,
		mapValue,
		mapValueType,
		mapValueFile string
	)
	cmd := &cobra.Command{
		Use:   "put [--name mapname | --key keyname | --value-type type | --value-file file | --value value]",
		Short: "Put value to map",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			var err error
			conf := cmd.Context().Value(config.HZCConfKey).(*hazelcast.Config)
			m, err := getMap(ctx, conf, mapName)
			if err != nil {
				return err
			}
			var normalizedValue interface{}
			if normalizedValue, err = normalizeMapValue(mapValue, mapValueFile, mapValueType); err != nil {
				return err
			}
			_, err = m.Put(ctx, mapKey, normalizedValue)
			if err == nil {
				return err
			}
			cmd.Printf("Cannot put value for key %s to map %s\n", mapKey, mapName)
			isCloudCluster := conf.Cluster.Cloud.Enabled
			if networkErrMsg, handled := internal.TranslateNetworkError(err, isCloudCluster); handled {
				err = hzcerror.NewLoggableError(err, networkErrMsg)
			}
			return err
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName)
	decorateCommandWithKeyFlags(cmd, &mapKey)
	decorateCommandWithValueFlags(cmd, &mapValue, &mapValueFile, &mapValueType)
	return cmd
}

func normalizeMapValue(v, vFile, vType string) (interface{}, error) {
	var valueStr string
	var err error
	switch {
	case v != "" && vFile != "":
		return nil, hzcerror.NewLoggableError(nil, "Only one of --value and --value-file must be specified")
	case v != "":
		valueStr = v
	case vFile != "":
		if valueStr, err = loadValueFile(vFile); err != nil {
			err = hzcerror.NewLoggableError(err, "Cannot load the value file. Make sure file exists and process has correct access rights")
		}
	default:
		err = hzcerror.NewLoggableError(nil, "One of the value flags (--value or --value-file) must be set")
	}
	if err != nil {
		return nil, err
	}
	switch vType {
	case internal.TypeString:
		return valueStr, nil
	case internal.TypeJSON:
		return serialization.JSON(valueStr), nil
	}
	return nil, hzcerror.NewLoggableError(nil, "Provided value type parameter (%s) is not a known type. Provide either 'string' or 'json'", vType)
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
