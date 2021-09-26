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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var (
	newState              string
	clusterChangeStateCmd = &cobra.Command{
		Use:   "change-state [--state new-state]",
		Short: "change state of the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			defer internal.ErrorRecover()
			config, err := internal.MakeConfig(cmd)
			if err != nil {
				return
			}
			result, err := internal.CallClusterOperation(config, "change-state", &newState)
			if err != nil {
				return
			}
			fmt.Println(*result)
		},
	}
)

func init() {
	clusterChangeStateCmd.Flags().StringVarP(&newState, "state", "s", "", "new state of the cluster")
	clusterChangeStateCmd.MarkFlagRequired("state")
	clusterChangeStateCmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"active", "no_migration", "frozen", "passive"}, cobra.ShellCompDirectiveDefault
	})
}
