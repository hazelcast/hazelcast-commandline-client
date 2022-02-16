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
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/c-bata/go-prompt"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/commands"
	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
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
			errStr := HandleError(err)
			rootCmd.Println(errStr)
			return
		},
	}
	rootCmd.Println("Connecting to the cluster ...")
	if _, err := internal.ConnectToCluster(ctx, conf); err != nil {
		rootCmd.Printf("Error: %s\n", err)
		return
	}
	var flagsToExclude []string
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flagsToExclude = append(flagsToExclude, flag.Name)
		// Mark hidden to exclude from help text in interactive mode.
		flag.Hidden = true
	})
	flagsToExclude = append(flagsToExclude, "help")
	p.FlagsToExclude = flagsToExclude
	rootCmd.Example = `> map put -k key -n myMap -v someValue
> map get -k key -m myMap
> cluster version`
	rootCmd.Use = ""
	p.Run(ctx)
}

func InitRootCmd() (*cobra.Command, *config.GlobalFlagValues) {
	rootCmd, persistentFlags := commands.NewRoot()
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetOut(os.Stdout)
	return rootCmd, persistentFlags
}

func getConfigWithFlags(rootCmd *cobra.Command, programArgs []string, globalFlagValues *config.GlobalFlagValues) (*hazelcast.Config, error) {
	// parse global persistent flags
	subCmd, flags, _ := rootCmd.Find(programArgs)
	// fall back to cmd.Help, even if there is error
	_ = subCmd.ParseFlags(flags)
	// initialize config from file
	conf, err := config.Get(*globalFlagValues)
	return conf, err
}

func HandleError(err error) string {
	errStr := fmt.Sprintf("Unknown Error: %s\n"+
		"Use \"hzc [command] --help\" for more information about a command.", err.Error())
	var loggable hzcerror.LoggableError
	if errors.As(err, &loggable) {
		errStr = fmt.Sprintf("Error: %s\n", loggable.VerboseError())
	}
	return errStr
}

func RunCmd(ctx context.Context, root *cobra.Command) error {
	ctx, cancel := context.WithCancel(ctx)
	handleInterrupt(ctx, cancel)
	return root.ExecuteContext(ctx)
}

func handleInterrupt(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
}
