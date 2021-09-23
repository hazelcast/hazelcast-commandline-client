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
	"log"
	"os"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var mapPutCmd = &cobra.Command{
	Use:   "put [--name mapname | --key keyname | --value-type type | --value-file file | --value value]",
	Short: "put to map",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.TODO()
		var err error
		var normalizedValue interface{}
		config, err := internal.MakeConfig(cmd)
		if err != nil {
			return err
		}
		m, err := getMap(config, mapName)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Fatal(internal.ErrConnectionTimeout)
			}
			return fmt.Errorf("error getting map %s: %w", mapName, err)
		}
		if mapKey == "" {
			return internal.ErrMapKeyMissing
		}
		if normalizedValue, err = normalizeMapValue(); err != nil {
			return err
		}
		_, err = m.Put(ctx, mapKey, normalizedValue)
		if err != nil {
			return fmt.Errorf("error putting value for key %s from map %s: %w", mapKey, mapName, err)
		}
		return nil
	},
}

func normalizeMapValue() (interface{}, error) {
	var valueStr string
	var err error
	if mapValue != "" && mapValueFile != "" {
		return nil, internal.ErrMapValueAndFileMutuallyExclusive
	} else if mapValue != "" {
		valueStr = mapValue
	} else if mapValueFile != "" {
		if valueStr, err = loadValueFile(mapValueFile); err != nil {
			return nil, fmt.Errorf("error loading value: %w", err)
		}
	} else {
		return nil, internal.ErrMapValueMissing
	}
	switch mapValueType {
	case internal.TypeString:
		return valueStr, nil
	case internal.TypeJSON:
		return serialization.JSON(valueStr), nil
	}
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
