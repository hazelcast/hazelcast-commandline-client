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

const invocationOnCloudErrorMessage = "Cluster operations on cloud are not supported. Checkout https://github.com/hazelcast/hazelcast-cloud-cli for cluster management on cloud."

var ClusterCmd = &cobra.Command{
	Use:   "cluster {get-state | change-state | shutdown | query} [--state new-state]",
	Short: "administrative cluster operations",
	Long:  `administrative cluster operations which controls a Hazelcast cluster by manipulating its state and other features`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hzConf, err := internal.MakeConfig()
		if err != nil {
			return err
		}
		// check if it is cloud invocation
		if hzConf.Cluster.Cloud.Token != "" {
			fmt.Println(invocationOnCloudErrorMessage)
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	ClusterCmd.AddCommand(clusterGetStateCmd)
	ClusterCmd.AddCommand(clusterChangeStateCmd)
	ClusterCmd.AddCommand(clusterShutdownCmd)
	ClusterCmd.AddCommand(clusterVersionCmd)
}
