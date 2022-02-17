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
	"strings"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const invocationOnCloudInfoMessage = "Cluster operations on cloud are not supported. Checkout https://github.com/hazelcast/hazelcast-cloud-cli for cluster management on cloud."

func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "cluster {get-state | change-state | shutdown | query} [--state new-state]",
		Short: "administrative cluster operations",
		Long:  `administrative cluster operations which controls a Hazelcast cluster by manipulating its state and other features`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			conf := config.FromContext(cmd.Context())
			// check if it is a cloud invocation
			if conf.Cluster.Cloud.Enabled {
				return hzcerror.NewLoggableError(nil, invocationOnCloudInfoMessage)
			}
			return nil
		},
	}
	subCmds := []struct {
		command string
		info    string
	}{
		{
			command: "shutdown",
			info:    "shuts down the cluster",
		},
		{
			command: "version",
			info:    "retrieve information from the cluster",
		},
		{
			command: "get-state",
			info:    "get state of the cluster",
		},
	}
	for _, sc := range subCmds {
		cmd.AddCommand(&cobra.Command{
			Use:   sc.command,
			Short: sc.info,
			RunE: func(cmd *cobra.Command, args []string) error {
				defer internal.ErrorRecover()
				conf := config.FromContext(cmd.Context())
				result, err := internal.CallClusterOperation(conf, sc.command)
				if err != nil {
					return err
				}
				cmd.Println(*result)
				return nil
			},
		})
	}
	// adding this explicitly, since it is a bit different from the rest
	cmd.AddCommand(NewChangeState())
	return &cmd
}

var states = []string{"active", "no_migration", "frozen", "passive"}

func NewChangeState() *cobra.Command {
	// monitored flag variable
	var newState string
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("change-state [--state [%s]]", strings.Join(states, ",")),
		Short: "change state of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			defer internal.ErrorRecover()
			conf := config.FromContext(cmd.Context())
			result, err := internal.CallClusterOperationWithState(conf, "change-state", &newState)
			if err != nil {
				return err
			}
			cmd.Println(*result)
			return nil
		},
	}
	cmd.Flags().StringVarP(&newState, "state", "s", "", fmt.Sprintf("new state of the cluster: %s", strings.Join(states, ",")))
	cmd.MarkFlagRequired("state")
	cmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return states, cobra.ShellCompDirectiveDefault
	})
	return cmd
}
