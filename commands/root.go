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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	clusterCmd "github.com/hazelcast/hazelcast-commandline-client/commands/cluster"
	sqlCmd "github.com/hazelcast/hazelcast-commandline-client/commands/sql"
	fakeDoor "github.com/hazelcast/hazelcast-commandline-client/commands/types/fakedoor"
	mapCmd "github.com/hazelcast/hazelcast-commandline-client/commands/types/map"
	"github.com/hazelcast/hazelcast-commandline-client/config"
)

// NewRoot initializes root command for non-interactive mode
func NewRoot() (*cobra.Command, *config.GlobalFlagValues) {
	var flags config.GlobalFlagValues
	root := &cobra.Command{
		Use:   "hzc {cluster | map | sql | help} [--address address | --cloud-token token | --cluster-name name | --config config]",
		Short: "Hazelcast command-line client",
		Long:  "Hazelcast command-line client connects your command-line to a Hazelcast cluster",
		Example: "`hzc` - starts an interactive shell 🚀\n" +
			"`hzc map --name my-map put --key hello --value world` - put entry into map directly\n" +
			"`hzc help` - print help",
		// Handle errors explicitly
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Make sure the command receives non-nil configuration
			conf := config.FromContext(cmd.Context())
			if conf == nil {
				return errors.New("missing configuration")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	// This is used to generate completion scripts
	if mode := os.Getenv("MODE"); strings.EqualFold(mode, "dev") {
		root.CompletionOptions.DisableDefaultCmd = false
	}
	assignPersistentFlags(root, &flags)
	root.AddCommand(subCommands()...)
	return root, &flags
}

func InitRootCmd() (*cobra.Command, *config.GlobalFlagValues) {
	rootCmd, persistentFlags := NewRoot()
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetOut(os.Stdout)
	return rootCmd, persistentFlags
}

func subCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		clusterCmd.New(),
		mapCmd.New(),
		sqlCmd.New(),
	}
	fds := []fakeDoor.FakeDoor{
		{Name: "list", IssueNum: 48},
		{Name: "queue", IssueNum: 49},
		{Name: "multimap", IssueNum: 50},
		{Name: "replicatedmap", IssueNum: 51},
		{Name: "set", IssueNum: 52},
		{Name: "topic", IssueNum: 53},
	}
	for _, fd := range fds {
		cmds = append(cmds, fakeDoor.MakeFakeCommand(fd))
	}
	return cmds
}

// assignPersistentFlags assigns top level flags to command
func assignPersistentFlags(cmd *cobra.Command, flags *config.GlobalFlagValues) {
	cmd.PersistentFlags().StringVarP(&flags.CfgFile, "config", "c", config.DefaultConfigPath(), fmt.Sprintf("config file, only supports yaml for now"))
	cmd.PersistentFlags().StringVarP(&flags.Address, "address", "a", "", fmt.Sprintf("addresses of the instances in the cluster (default is %s).", config.DefaultClusterAddress))
	cmd.PersistentFlags().StringVarP(&flags.Cluster, "cluster-name", "", "", fmt.Sprintf("name of the cluster that contains the instances (default is %s).", config.DefaultClusterName))
	cmd.PersistentFlags().StringVar(&flags.Token, "cloud-token", "", "your Hazelcast Cloud token.")
	cmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "", false, "verbose output.")
}
