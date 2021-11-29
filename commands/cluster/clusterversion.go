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

var clusterVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "retrieve information from the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		defer internal.ErrorRecover()
		config, err := internal.MakeConfig()
		//TODO error look like unhandled although it is handled in MakeConfig. Find a better approach
		if err != nil {
			return
		}
		// check if it is cloud invocation
		if config.Cluster.Cloud.Token != "" {
			fmt.Println(invocationOnCloudErrorMessage)
			return
		}
		result, err := internal.CallClusterOperation(config, "version")
		if err != nil {
			return
		}
		fmt.Println(*result)
	},
}

func init() {
}
