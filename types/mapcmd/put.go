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
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	fds "github.com/hazelcast/hazelcast-commandline-client/internal/flagdecorators"

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
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s", mapKey, mapKeyType)
			}
			var (
				uTTL,
				uMaxIdle *internal.UserDuration
			)
			if ttl.Seconds() != 0 {
				uTTL = &internal.UserDuration{Duration: ttl, DurType: internal.TTL}
				if err = uTTL.Validate(); err != nil {
					return hzcerrors.NewLoggableError(err, "ttl is invalid")
				}
			}
			if maxIdle.Seconds() != 0 {
				uMaxIdle = &internal.UserDuration{Duration: maxIdle, DurType: internal.MaxIdle}
				if err = uMaxIdle.Validate(); err != nil {
					return hzcerrors.NewLoggableError(err, "max-idle is invalid")
				}
			}
			var normalizedValue interface{}
			if normalizedValue, err = normalizeMapValue(mapValue, mapValueFile, mapValueType); err != nil {
				return err
			}
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			switch {
			case uTTL != nil && uMaxIdle != nil:
				_, err = m.PutWithTTLAndMaxIdle(ctx, key, normalizedValue, uTTL.Duration, uMaxIdle.Duration)
			case uTTL != nil:
				_, err = m.PutWithTTL(ctx, key, normalizedValue, uTTL.Duration)
			case uMaxIdle != nil:
				_, err = m.PutWithMaxIdle(ctx, key, normalizedValue, uMaxIdle.Duration)
			default:
				_, err = m.Put(ctx, key, normalizedValue)
			}
			if err != nil {
				var handled bool
				handled, err = cloudcb(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot put given entry to the map %s", mapName)
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyFlags(cmd, &mapKey, &mapKeyType, true, "key of the entry")
	decorateCommandWithValueFlags(cmd, &mapValue, &mapValueFile, &mapValueType)
	fds.DecorateCommandWithTTL(cmd, &ttl, false, "ttl value of the entry")
	fds.DecorateCommandWithMaxIdle(cmd, &maxIdle, false, "max-idle value of the entry")
	return cmd
}
