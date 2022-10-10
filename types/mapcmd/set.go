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
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const MapSetExample = `  # Set puts key, value pair to map. The unit for ttl/max-idle is one of (ns,us,ms,s,m,h)
  map set --key-type string --key hello --value-type float32 --value 19.94 --name myMap --ttl 1300ms --max-idle 1400ms`

func NewSet(config *hazelcast.Config) *cobra.Command {
	var (
		mapName,
		mapKey,
		mapKeyType,
		mapValue,
		mapValueType,
		mapValueFile string
	)
	var (
		ttl,
		maxIdle time.Duration
	)
	cmd := &cobra.Command{
		Use:     "set [--name mapname | --key keyname | --value-type type | {--value-file file | --value value} | --ttl ttl | --max-idle max-idle]",
		Short:   "Set key and corresponding value on map",
		Example: MapSetExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s, %s", mapKey, mapKeyType, err)
			}
			var ttlIsSet, maxIdleIsSet bool
			if ttlIsSet = ttl.Seconds() != 0; ttlIsSet {
				if err = validateTTL(ttl); err != nil {
					return hzcerrors.NewLoggableError(err, "ttl is invalid")
				}
			}
			if maxIdleIsSet = maxIdle.Seconds() != 0; maxIdleIsSet {
				if err = isNegativeSecond(maxIdle); err != nil {
					return hzcerrors.NewLoggableError(err, "max-idle is invalid")
				}
				maxIdleIsSet = true
			}
			var normalizedValue interface{}
			if normalizedValue, err = normalizeMapValue(mapValue, mapValueFile, mapValueType); err != nil {
				return err
			}
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			switch {
			case ttlIsSet && maxIdleIsSet:
				err = m.SetWithTTLAndMaxIdle(cmd.Context(), key, normalizedValue, ttl, maxIdle)
			case ttlIsSet:
				err = m.SetWithTTL(cmd.Context(), key, normalizedValue, ttl)
			case maxIdleIsSet:
				hackForUnSet := time.Duration(0)
				err = m.SetWithTTLAndMaxIdle(cmd.Context(), key, normalizedValue, hackForUnSet, maxIdle)
			default:
				err = m.Set(cmd.Context(), key, normalizedValue)
			}
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot put given entry to the map %s", mapName)
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry")
	decorateCommandWithMapKeyTypeFlags(cmd, &mapKeyType, false)
	decorateCommandWithValueFlags(cmd, &mapValue, &mapValueFile)
	decorateCommandWithMapValueTypeFlags(cmd, &mapValueType, false)
	decorateCommandWithTTL(cmd, &ttl, false, "ttl value of the entry")
	decorateCommandWithMaxIdle(cmd, &maxIdle, false, "max-idle value of the entry")
	return cmd
}
