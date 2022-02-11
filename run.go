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
package main

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
	persister "github.com/hazelcast/hazelcast-commandline-client/internal/persistency"
)

func IsInteractiveCall(rootCmd *cobra.Command, args []string) bool {
	cmd, flags, err := rootCmd.Find(args)
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
	if cmd == rootCmd {
		return true
	}
	return false
}

func RunCmdInteractively(ctx context.Context, rootCmd *cobra.Command, conf *hazelcast.Config) {
	namePersister := persister.NewNamePersister()
	ctx = context.WithValue(ctx, "persister", namePersister)
	var p = &cobraprompt.CobraPrompt{
		RootCmd:                  rootCmd,
		ShowHelpCommandAndFlags:  true,
		ShowHiddenFlags:          true,
		SuggestFlagsWithoutDash:  true,
		DisableCompletionCommand: true,
		AddDefaultExitCommand:    true,
		GoPromptOptions: []prompt.Option{
			prompt.OptionTitle("Hazelcast Client"),
			prompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
				return fmt.Sprintf("hzc %s@%s> ", config.GetClusterAddress(conf), conf.Cluster.Name), true
			}),
			prompt.OptionMaxSuggestion(10),
			prompt.OptionCompletionOnDown(),
		},
		OnErrorFunc: func(err error) {
			HandleError(rootCmd, err)
			return
		},
		Persister: namePersister,
	}
	rootCmd.Println("Connecting to the cluster ...")
	if _, err := internal.ConnectToCluster(ctx, conf); err != nil {
		rootCmd.Printf("Error: %s\n", err)
		return
	}
	var flagsToExclude []string
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flagsToExclude = append(flagsToExclude, flag.Name)
	})
	flagsToExclude = append(flagsToExclude, "help")
	p.FlagsToExclude = flagsToExclude
	rootCmd.Example = `> map put -k key -n myMap -v someValue
> map get -k key -m myMap
> cluster version`
	rootCmd.Use = ""
	p.Run(ctx)
}
