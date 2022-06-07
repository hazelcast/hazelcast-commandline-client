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
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const MapUseExample = `  hzc map use m1    # sets the default map name to m1 unless set explicitly
  hzc map get --key k1    # "--name m1" is inferred
  hzc map use --reset	  # resets the behaviour`

func NewUse() *cobra.Command {
	cmd := &cobra.Command{
		Use:     `use [map-name | --reset]`,
		Short:   "sets the default map name (interactive-mode only)",
		Example: MapUseExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			persister := internal.PersistedNamesFromContext(cmd.Context())
			if cmd.Flags().Changed("reset") {
				delete(persister, "map")
				return nil
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			if len(args) > 1 {
				cmd.Println("Provide map name between \"\" quotes if it contains white space")
				return nil
			}
			persister["map"] = args[0]
			return nil
		},
	}
	_ = cmd.Flags().BoolP(MapResetFlag, "", false, "unset default name for map")
	return cmd
}
