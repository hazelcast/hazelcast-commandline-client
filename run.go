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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/connection"
	"github.com/hazelcast/hazelcast-commandline-client/internal/file"
	goprompt "github.com/hazelcast/hazelcast-commandline-client/internal/go-prompt"
	"github.com/hazelcast/hazelcast-commandline-client/types/mapcmd"
)

const (
	ViridianCoordinatorURL       = "https://api.viridian.hazelcast.com"
	EnvHzCloudCoordinatorBaseURL = "HZ_CLOUD_COORDINATOR_BASE_URL"
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

func RunCmdInteractively(ctx context.Context, rootCmd *cobra.Command, cnfg *config.Config, noColor bool) {
	cmdHistoryPath := filepath.Join(file.HZCHomePath(), "history")
	exists, err := file.Exists(cmdHistoryPath)
	if err != nil {
		// todo log err once we have logging solution
	}
	if !exists {
		if err := file.CreateMissingDirsAndFileWithRWPerms(cmdHistoryPath, []byte{}); err != nil {
			// todo log err once we have logging solution
		}
	}
	hConfig := &cnfg.Hazelcast
	namePersister := make(map[string]string)
	var p = &cobraprompt.CobraPrompt{
		ShowHelpCommandAndFlags:  true,
		ShowHiddenFlags:          true,
		SuggestFlagsWithoutDash:  true,
		DisableCompletionCommand: true,
		DisableSuggestions:       cnfg.NoAutocompletion,
		NoColor:                  noColor,
		AddDefaultExitCommand:    true,
		GoPromptOptions: []goprompt.Option{
			goprompt.OptionTitle("Hazelcast Client"),
			goprompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
				var b strings.Builder
				for k, v := range namePersister {
					b.WriteString(fmt.Sprintf("&%c:%s", k[0], v))
				}
				return fmt.Sprintf("hzc %s@%s%s> ", config.GetClusterAddress(hConfig), hConfig.Cluster.Name, b.String()), true
			}),
			goprompt.OptionMaxSuggestion(10),
			goprompt.OptionCompletionOnDown(),
		},
		OnErrorFunc: func(err error) {
			errStr := HandleError(err)
			cnfg.Logger.Println(errStr)
			return
		},
		Persister: namePersister,
	}
	if _, err = connection.ConnectToClusterInteractive(ctx, hConfig); err != nil {
		// ignore error coming from the connection spinner
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
	rootCmd.Example = fmt.Sprintf("> %s\n> %s", mapcmd.MapPutExample, mapcmd.MapGetExample) + "\n> cluster version"
	rootCmd.Use = ""
	p.Run(ctx, rootCmd, hConfig, cmdHistoryPath)
	return
}

func updateConfigWithFlags(rootCmd *cobra.Command, cnfg *config.Config, programArgs []string, globalFlagValues *config.GlobalFlagValues) error {
	// parse global persistent flags
	subCmd, flags, _ := rootCmd.Find(programArgs)
	// fall back to cmd.Help, even if there is error
	_ = subCmd.ParseFlags(flags)
	// initialize config from file
	if err := config.ReadAndMergeWithFlags(globalFlagValues, cnfg); err != nil {
		return err
	}
	if cnfg.Hazelcast.Cluster.Cloud.Enabled {
		return setDefaultCoordinator()
	}
	return nil
}

func setDefaultCoordinator() error {
	if os.Getenv(EnvHzCloudCoordinatorBaseURL) != "" {
		return nil
	}
	// if not set assign Viridian
	if err := os.Setenv(EnvHzCloudCoordinatorBaseURL, ViridianCoordinatorURL); err != nil {
		return hzcerrors.NewLoggableError(err, "Can not assign Viridian as the default coordinator")
	}
	return nil
}

func HandleError(err error) string {
	errStr := fmt.Sprintf("Unknown Error: %s\n"+
		"Use \"hzc [command] --help\" for more information about a command.", err.Error())
	var loggable hzcerrors.LoggableError
	var flagErr hzcerrors.FlagError
	if errors.As(err, &loggable) {
		errStr = fmt.Sprintf("Error: %s\n", loggable.VerboseError())
	} else if errors.As(err, &flagErr) {
		errStr = fmt.Sprintf("Flag Error: %s\n", err.Error())
	}
	return errStr
}

func RunCmd(ctx context.Context, rootCmd *cobra.Command) error {
	p := make(map[string]string)
	ctx = internal.ContextWithPersistedNames(ctx, p)
	ctx, cancel := context.WithCancel(ctx)
	handleInterrupt(ctx, cancel)
	return rootCmd.ExecuteContext(ctx)
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
