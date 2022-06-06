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

const MapPutExample = `  # Put key, value pair to map.
  hzc put -n mapname -k k1 -v v1

  # Put key, value pair to map but type of the value is accepted as json data.
  hzc map put -n mapname -k k2 -v v2 -t json
  
  # Put key, value pair to map but the value is given through external file.
  hzc map put -n mapname -k k3 -f v3.txt
  
  # Put key, value pair to map but the value is given through external json file.
  hzc map put -n mapname -k k4 -f v4.json -t json
  
  # Put key, value pair to map with given ttl value
  hzc map put -n mapname -k k5 -v v5 --ttl 3ms
  
  # Put key, value pair to map with given ttl and maxidle values
  hzc map put -n mapname -k k1 -v v1 --ttl 3ms --max-idle 4ms
  
  # Put custom type key and value to map
  map put --key-type string --key hello --value-type float32 --value 19.94 --name myMap

  # TTL and Maxidle:
  ttl and maxidle values cannot be less than a second when it is converted to second from any time unit. Supported units are;
    - Nanosecond  (ns)
    - Microsecond (μs)
    - Millisecond (ms)
    - Second      (s)
    - Minute      (m)
    - Hour        (h)`

func NewPut(config *hazelcast.Config) *cobra.Command {
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
		Use:     "put [--name mapname | --key keyname | --value-type type | {--value-file file | --value value} | --ttl ttl | --max-idle max-idle]",
		Short:   "Put value to map",
		Example: MapPutExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s", mapKey, mapKeyType)
			}
			var (
				ttlE,
				maxIdleE bool
			)
			if ttl.Seconds() != 0 {
				if err = validateTTL(ttl); err != nil {
					return hzcerrors.NewLoggableError(err, "ttl is invalid")
				}
				ttlE = true
			}
			if maxIdle.Seconds() != 0 {
				if err = isNegativeSecond(maxIdle); err != nil {
					return hzcerrors.NewLoggableError(err, "max-idle is invalid")
				}
				maxIdleE = true
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
			case ttlE && maxIdleE:
				_, err = m.PutWithTTLAndMaxIdle(cmd.Context(), key, normalizedValue, ttl, maxIdle)
			case ttlE:
				_, err = m.PutWithTTL(cmd.Context(), key, normalizedValue, ttl)
			case maxIdleE:
				_, err = m.PutWithMaxIdle(cmd.Context(), key, normalizedValue, maxIdle)
			default:
				_, err = m.Put(cmd.Context(), key, normalizedValue)
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
