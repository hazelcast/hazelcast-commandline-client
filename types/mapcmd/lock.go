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

const MapLockExample = `  # Lock the given map.
  hzc map lock --key mapkey --name mapname --lease-time 10ms`

func NewLock(config *hazelcast.Config) *cobra.Command {
	var (
		mapName, mapKey, mapKeyType string
		leaseTime                   time.Duration
	)
	cmd := &cobra.Command{
		Use:     "lock --key mapkey --name mapname [--lease-time duration]",
		Short:   "Lock the map",
		Example: MapLockExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s, %s", mapKey, mapKeyType, err)
			}
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			if leaseTime.Milliseconds() != 0 {
				err = m.LockWithLease(cmd.Context(), key, leaseTime)
			} else {
				err = m.Lock(cmd.Context(), key)
			}
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot get the size of the map %s", mapName)
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry")
	decorateCommandWithMapKeyTypeFlags(cmd, &mapKeyType, false)
	decorateCommandWithLeaseTime(cmd, &leaseTime, false, "duration to hold the lock (default: indefinitely)")
	return cmd
}
