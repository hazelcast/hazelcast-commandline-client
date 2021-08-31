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
	"log"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/spf13/cobra"
)

var clusterGetStateCmd = &cobra.Command{
	Use:   "get-state",
	Short: "get state of the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		defer internal.ErrorRecover()
		config, err := internal.MakeConfig(cmd)
		if err != nil {
			log.Fatal(err)
		}
		result, err := internal.CallClusterOperation(config, "get-state", nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(*result)
	},
}

func init() {
}
