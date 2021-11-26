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
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clustercmd "github.com/hazelcast/hazelcast-commandline-client/commands/cluster"
	mapcmd "github.com/hazelcast/hazelcast-commandline-client/commands/types/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
)

var (
	RootCmd = &cobra.Command{
		Use:   "hzc {cluster | help | map} [--address address | --cloud-token token | --cluster-name name | --config config]",
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
	if config := internal.Configuration; config != nil {
		return fmt.Sprintf("hzc %s@%s> ", config.Cluster.Network.Addresses[0], config.Cluster.Name), true
	}
	return "hzc address@clusterName> ", true
}

var advancedPrompt = &cobraprompt.CobraPrompt{
	RootCmd:                  RootCmd,
	PersistFlagValues:        true,
	ShowHelpCommandAndFlags:  true,
	ShowHiddenFlags:          true,
	SuggestFlagsWithoutDash:  true,
	DisableCompletionCommand: true,
	AddDefaultExitCommand:    true,
	GoPromptOptions: []prompt.Option{
		prompt.OptionTitle("hazelcast commandline client"),
		prompt.OptionLivePrefix(addressAndClusterNamePrefix),
		prompt.OptionMaxSuggestion(10),
	},
	OnErrorFunc: func(err error) {
		RootCmd.PrintErrln(err)
		RootCmd.Help()
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

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func ExecuteInteractive() {
	RootCmd.RunE = nil                                  // disable help text on each new line
	if err := RootCmd.ParseFlags(os.Args); err != nil { // to parse global persistent flags
		log.Fatal(err)
	}
	conf, err := internal.MakeConfig()
	if err != nil { // initialize global config
		log.Fatal(err)
	}
	if _, err = internal.ConnectToCluster(context.Background(), conf.Clone()); err != nil {
		fmt.Println("Error creating the client: %w", err)
		return
	}
	var flagsToExclude []string
	RootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flagsToExclude = append(flagsToExclude, flag.Name)
	})
	advancedPrompt.FlagsToExclude = flagsToExclude
	RootCmd.SetHelpFunc(customHelp)
	RootCmd.SilenceUsage = true
	RootCmd.Example = "> map put -k key -m myMap -v someValue\n" +
		"> map get -k key -m myMap\n" +
		"> cluster version"
	RootCmd.Use = ""
	advancedPrompt.Run(context.Background())
}

func decorateRootCommand(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&internal.CfgFile, "config", "c", internal.DefautConfigPath(), fmt.Sprintf("config file, only supports yaml for now (default path is %s).", internal.DefautConfigPath()))
	cmd.PersistentFlags().StringVarP(&internal.Address, "address", "a", "", fmt.Sprintf("addresses of the instances in the cluster (default is %s).", internal.DefaultClusterAddress))
	cmd.PersistentFlags().StringVarP(&internal.Cluster, "cluster-name", "", "", fmt.Sprintf("name of the cluster that contains the instances (default is %s).", internal.DefaultClusterName))
	cmd.PersistentFlags().StringVar(&internal.Token, "cloud-token", "", "your Hazelcast Cloud token.")
	cmd.CompletionOptions.DisableDefaultCmd = true

	if mode := os.Getenv("MODE"); strings.EqualFold(mode, "dev") { // This is used to generate completion scripts
		cmd.CompletionOptions.DisableDefaultCmd = false
	}
}

func customHelp(command *cobra.Command, args []string) {
	if len(args) == 0 && command.Short == "Hazelcast command-line client" { // do not show help for root command
		return
	}
	command.HelpFunc()
	command.SetHelpFunc(nil)
	command.Help()
	command.SetHelpFunc(customHelp)
}

func initRootCommand(rootCmd *cobra.Command) {
	decorateRootCommand(rootCmd)
	rootCmd.AddCommand(clustercmd.ClusterCmd, mapcmd.MapCmd)
}

func init() {
	initRootCommand(RootCmd)
}
