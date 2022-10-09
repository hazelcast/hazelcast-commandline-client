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

package clustercmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const invocationOnCloudInfoMessage = "Cluster operations on cloud are not supported. Checkout https://github.com/hazelcast/hazelcast-cloud-cli for cluster management on cloud."

func New(config *hazelcast.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "cluster {get-state | change-state | shutdown | monitor | version} [--state new-state]",
		Short: "Administrative cluster operations",
		Long:  `Administrative cluster operations which controls a Hazelcast cluster by manipulating its state and other features`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// check if it is a cloud invocation
			if config.Cluster.Cloud.Enabled {
				return hzcerrors.NewLoggableError(nil, invocationOnCloudInfoMessage)
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
		// copy to use it in the inner func
		sc := sc
		cmd.AddCommand(&cobra.Command{
			Use:   sc.command,
			Short: sc.info,
			RunE: func(cmd *cobra.Command, args []string) error {
				defer hzcerrors.ErrorRecover()
				result, err := internal.CallClusterOperation(config, sc.command)
				if err != nil {
					return err
				}
				cmd.Println(*result)
				return nil
			},
		})
	}
	// adding this explicitly, since it is a bit different from the rest
	cmd.AddCommand(NewChangeState(config))
	cmd.AddCommand(NewMonitorCmd(config))
	return &cmd
}

var states = []string{"active", "no_migration", "frozen", "passive"}

func NewChangeState(config *hazelcast.Config) *cobra.Command {
	// monitored flag variable
	var newState string
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("change-state [--state [%s]]", strings.Join(states, ",")),
		Short: "Change state of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			defer hzcerrors.ErrorRecover()
			result, err := internal.CallClusterOperationWithState(config, "change-state", &newState)
			if err != nil {
				return err
			}
			fmt.Println(*result)
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

func NewMonitorCmd(config *hazelcast.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Monitor the cluster for changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			var ci *hazelcast.ClientInternal
			config.AddLifecycleListener(func(event hazelcast.LifecycleStateChanged) {
				state := strings.ReplaceAll(strings.ToUpper(event.State.String()), " ", "_")
				t := time.Now().Format("2006-01-02 15:04:05")
				clusterID := "-"
				if ci != nil {
					clusterID = ci.ClusterID().String()
				}
				fmt.Printf("%s\t%-19s\t%s\n", t, state, clusterID)
			})
			config.AddMembershipListener(func(event cluster.MembershipStateChanged) {
				state := strings.ToUpper(event.State.String())
				m := event.Member
				t := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("%s\t%-19s\t%s\t%s\t%s\n", t, state, m.Address, m.Version, m.UUID)
			})
			client, err := internal.ConnectToCluster(cmd.Context(), config)
			if err != nil {
				return err
			}
			ci = hazelcast.NewClientInternal(client)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			internal.WaitTerminate(ctx, cancel)
			return nil
		},
	}
}
