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

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/types/flagdecorators"
)

func NewPut(config *hazelcast.Config) (*cobra.Command, error) {
	var (
		mapName,
		mapKey,
		mapValue,
		mapValueType,
		mapValueFile string
	)
	var (
		ttl,
		maxIdle time.Duration
	)
	cmd := &cobra.Command{
		Use:   "put [--name mapname | --key keyname | --value-type type | --value-file file | --value value | --ttl ttl | --max-idle max-idle]",
		Short: "Put value to specified map",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			var err error
			var (
				uTTL,
				uMaxIdle *internal.UserDuration
			)
			if ttl.Seconds() != 0 {
				uTTL = &internal.UserDuration{Duration: ttl, DurType: internal.TTL}
				if err = uTTL.Validate(); err != nil {
					return hzcerror.NewLoggableError(err, "ttl is invalid")
				}
			}
			if maxIdle.Seconds() != 0 {
				uMaxIdle = &internal.UserDuration{Duration: maxIdle, DurType: internal.MaxIdle}
				if err = uMaxIdle.Validate(); err != nil {
					return hzcerror.NewLoggableError(err, "max-idle is invalid")
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
				_, err = m.PutWithTTLAndMaxIdle(ctx, mapKey, normalizedValue, uTTL.Duration, uMaxIdle.Duration)
			case uTTL != nil:
				_, err = m.PutWithTTL(ctx, mapKey, normalizedValue, uTTL.Duration)
			case uMaxIdle != nil:
				_, err = m.PutWithMaxIdle(ctx, mapKey, normalizedValue, uMaxIdle.Duration)
			default:
				_, err = m.Put(ctx, mapKey, normalizedValue)
			}
			if err != nil {
				var handled bool
				handled, err = cloudcb(err, config)
				if handled {
					return err
				}
				return hzcerror.NewLoggableError(err, "Cannot put given entry to the map %s", mapName)
			}
			return nil
		},
	}
	if err := decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithValueFlags(cmd, &mapValue, &mapValueFile, &mapValueType); err != nil {
		return nil, err
	}
	if err := flagdecorators.DecorateCommandWithTTL(cmd, &ttl, false, "ttl value of the entry"); err != nil {
		return nil, err
	}
	if err := flagdecorators.DecorateCommandWithMaxIdle(cmd, &maxIdle, false, "max-idle value of the entry"); err != nil {
		return nil, err
	}
	return cmd, nil
}
