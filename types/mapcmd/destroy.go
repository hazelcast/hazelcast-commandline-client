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
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

const MapDestroyExample = `  # Destroy the given map.
  hzc map destroy --name mapname`

func NewDestroy(config *hazelcast.Config) *cobra.Command {
	var mapName string
	cmd := &cobra.Command{
		Use:     "destroy --name mapname",
		Short:   "Destroy the map",
		Example: MapDestroyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getMap(cmd.Context(), config, mapName)
			if err != nil {
				fmt.Println("get map err ")
				return err
			}
			err = m.Destroy(cmd.Context())
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					fmt.Println("handled")
					return err
				}
				fmt.Println("normal err")
				return hzcerrors.NewLoggableError(err, "Cannot get the size of the map %s", mapName)
			}
			fmt.Println("normal return")
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	return cmd
}
