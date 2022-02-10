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
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clusterCmd "github.com/hazelcast/hazelcast-commandline-client/commands/cluster"
	fakeDoor "github.com/hazelcast/hazelcast-commandline-client/commands/types/fakedoor"
	mapCmd "github.com/hazelcast/hazelcast-commandline-client/commands/types/map"
	sqlCmd "github.com/hazelcast/hazelcast-commandline-client/commands/sql"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
)

var (
	RootCmd = &cobra.Command{
		Use:   "hzc {cluster | map | sql | help} [--address address | --cloud-token token | --cluster-name name | --config config]",
		Short: "Hazelcast command-line client",
		Long:  "Hazelcast command-line client connects your command-line to a Hazelcast cluster",
		Example: "`hzc map --name my-map put --key hello --value world` - put entry into map directly\n" +
			"`hzc help` - print help",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := internal.MakeConfig(); err != nil {
				return err
			}
			return cmd.Help()
		},
	}
)

func addressAndClusterNamePrefix() (prefix string, useLive bool) {
	return fmt.Sprintf("hzc %s@%s> ", internal.Configuration.Cluster.Network.Addresses[0], internal.Configuration.Cluster.Name), true
}

var advancedPrompt = &cobraprompt.CobraPrompt{
	RootCmd:                  RootCmd,
	ShowHelpCommandAndFlags:  true,
	ShowHiddenFlags:          true,
	SuggestFlagsWithoutDash:  true,
	DisableCompletionCommand: true,
	AddDefaultExitCommand:    true,
	GoPromptOptions: []prompt.Option{
		prompt.OptionTitle("Hazelcast Client"),
		prompt.OptionLivePrefix(addressAndClusterNamePrefix),
		prompt.OptionMaxSuggestion(10),
		prompt.OptionCompletionOnDown(),
	},
	OnErrorFunc: func(err error) {
		// handle error noop to prevent application from crashing
		return
	},
}

func IsInteractiveCall() bool {
	cmd, flags, err := RootCmd.Find(os.Args[1:])
	if err != nil {
		return false
	}
	for _, flag := range flags {
		if flag == "--help" || flag == "-h" {
			return false
		}
	}
	if cmd.Name() == "help" {
		return false
	}
	if cmd == RootCmd {
		return true
	}
	return false
}

func Execute(ctx context.Context) {
	cmdCtx, cmdCancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		select {
		case s := <-c:
			fmt.Println("signal received", s)
			cmdCancel()
		}
	}()
	if err := RootCmd.ExecuteContext(cmdCtx); err != nil {
		log.Fatal(err)
	}
}

func ExecuteInteractive(ctx context.Context) {
	cobraprompt.RegisterPersistFlag(RootCmd)
	// parse global persistent flags
	if err := RootCmd.ParseFlags(os.Args); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connecting to the cluster ...")
	if _, err := internal.ConnectToCluster(ctx); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	var flagsToExclude []string
	RootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flagsToExclude = append(flagsToExclude, flag.Name)
	})
	flagsToExclude = append(flagsToExclude, "help")
	advancedPrompt.FlagsToExclude = flagsToExclude
	RootCmd.Example = `> map put -k key -n myMap -v someValue
> map get -k key -m myMap
> cluster version`
	RootCmd.Use = ""
	advancedPrompt.Run(ctx)
}

func decorateRootCommand(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&internal.CfgFile, "config", "c", internal.DefaultConfigPath(), fmt.Sprintf("config file, only supports yaml for now"))
	cmd.PersistentFlags().StringVarP(&internal.Address, "address", "a", "", fmt.Sprintf("addresses of the instances in the cluster (default is %s).", internal.DefaultClusterAddress))
	cmd.PersistentFlags().StringVarP(&internal.Cluster, "cluster-name", "", "", fmt.Sprintf("name of the cluster that contains the instances (default is %s).", internal.DefaultClusterName))
	cmd.PersistentFlags().StringVar(&internal.Token, "cloud-token", "", "your Hazelcast Cloud token.")
	cmd.PersistentFlags().BoolVarP(&internal.Verbose, "verbose", "v", false, "verbose output.")
	cmd.CompletionOptions.DisableDefaultCmd = true
	// This is used to generate completion scripts
	if mode := os.Getenv("MODE"); strings.EqualFold(mode, "dev") {
		cmd.CompletionOptions.DisableDefaultCmd = false
	}
}

func subCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		clusterCmd.ClusterCmd,
		mapCmd.MapCmd,
		sqlCmd.SqlCmd,
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

func initRootCommand(rootCmd *cobra.Command) {
	decorateRootCommand(rootCmd)
	rootCmd.AddCommand(subCommands()...)
}

func init() {
	initRootCommand(RootCmd)
}
