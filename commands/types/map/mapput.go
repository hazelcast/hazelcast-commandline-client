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
package commands

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var mapPutCmd = &cobra.Command{
	Use:   "put [--name mapname | --key keyname | --value-type type | --value-file file | --value value]",
	Short: "put to map",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		var err error
		var normalizedValue interface{}
		config, err := internal.MakeConfig(cmd)
		if err != nil {
			return
		}
		m, err := getMap(config, mapName)
		if err != nil {
			return
		}
		if normalizedValue, err = normalizeMapValue(); err != nil {
			return
		}
		_, err = m.Put(ctx, mapKey, normalizedValue)
		if err != nil {
			fmt.Printf("error putting value for key %s from map %s\n", mapKey, mapName)
			return
		}
		return
	},
}

func normalizeMapValue() (interface{}, error) {
	var valueStr string
	var err error
	if mapValue != "" && mapValueFile != "" {
		fmt.Println("Error: Only one of --value and --value-file must be specified")
		return nil, errors.New("only one of --value and --value-file must be specified")
	} else if mapValue != "" {
		valueStr = mapValue
	} else if mapValueFile != "" {
		if valueStr, err = loadValueFile(mapValueFile); err != nil {
			fmt.Println("Error: Cannot load the value file. Please make sure file exists and process has correct access rights")
			return nil, fmt.Errorf("error loading value: %w", err)
		}
	} else {
		fmt.Println("Error: One of the value flag must be set")
		return nil, errors.New("map value is required")
	}
	switch mapValueType {
	case internal.TypeString:
		return valueStr, nil
	case internal.TypeJSON:
		return serialization.JSON(valueStr), nil
	}
	fmt.Println("Error: Provided value type parameter is not a known type. Please provide either 'string' or 'json'")
	return nil, fmt.Errorf("%s is not a known value type", mapValueType)
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

func init() {
	decorateCommandWithMapNameFlags(mapPutCmd)
	decorateCommandWithKeyFlags(mapPutCmd)
	decorateCommandWithValueFlags(mapPutCmd)
}
