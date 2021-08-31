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

	clusterDir "github.com/hazelcast/hazelcast-commandline-client/commands/cluster"
	mapDir "github.com/hazelcast/hazelcast-commandline-client/commands/types/map"
	"github.com/spf13/cobra"
)

const DefaultConfigFile = ".hzc.yaml"

var (
	cfgFile   string
	addresses string
	cluster   string
	token     string
	rootCmd   = &cobra.Command{
		Use:   "hz-cli {cluster | help | map} [--address address | --cloud-token token | --cluster-name name | --config config]",
		Short: "Hazelcast command-line client",
		Long:  "Hazelcast command-line client connects your command-line to a Hazelcast cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func decorateRootCommand(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", fmt.Sprintf("config file (default is $HOME/%s)", DefaultConfigFile))
	cmd.PersistentFlags().StringVarP(&addresses, "address", "a", "", "addresses of the instances in the cluster.")
	cmd.PersistentFlags().StringVarP(&cluster, "cluster-name", "n", "", "name of the cluster that contains the instances.")
	cmd.PersistentFlags().StringVar(&token, "cloud-token", "", "your Hazelcast Cloud token.")
	cmd.CompletionOptions.DisableDefaultCmd = true
}

func init() {
	decorateRootCommand(rootCmd)
	rootCmd.AddCommand(clusterDir.ClusterCmd)
	rootCmd.AddCommand(mapDir.MapCmd)
}
