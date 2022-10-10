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

const MapTryLockExample = `  # Try to lock the specified key of the specified map. Prints "unsuccessful" if not successful.
  hzc map trylock --key mapkey --name mapname --timeout 10ms --lease-time 2m`

func NewTryLock(config *hazelcast.Config) *cobra.Command {
	var (
		mapName, mapKey, mapKeyType string
		timeout, leaseTime          time.Duration
	)
	cmd := &cobra.Command{
		Use:     "trylock --key mapkey --name mapname [--lease-time duration] [--timeout duration]",
		Short:   "trylock the map",
		Example: MapTryLockExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s, %s", mapKey, mapKeyType, err)
			}
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				return err
			}
			var success bool
			ctx := cmd.Context()
			timeoutSet, leaseSet := timeout.Milliseconds() != 0, leaseTime.Milliseconds() != 0
			if timeoutSet && leaseSet {
				success, err = m.TryLockWithLeaseAndTimeout(ctx, key, leaseTime, timeout)
			} else if timeoutSet {
				success, err = m.TryLockWithTimeout(ctx, key, timeout)
			} else if leaseSet {
				success, err = m.TryLockWithLease(ctx, key, leaseTime)
			} else {
				success, err = m.TryLock(ctx, key)
			}
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Can not do tryLock operation on the map %s", mapName)
			}
			if !success {
				cmd.Println("unsuccessful")
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry")
	decorateCommandWithMapKeyTypeFlags(cmd, &mapKeyType, false)
	decorateCommandWithTimeout(cmd, &timeout, false, "duration to wait for the lock to be available")
	decorateCommandWithLeaseTime(cmd, &leaseTime, false, "duration to hold the lock (default: indefinitely)")
	return cmd
}
