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
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/file"
	goprompt "github.com/hazelcast/hazelcast-commandline-client/internal/go-prompt"
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

func RunCmdInteractively(ctx context.Context, rootCmd *cobra.Command, cnfg *hazelcast.Config) {
	cmdHistoryPath := filepath.Join(file.HZCHomePath(), ".hzc_history")
	exists, err := file.FileExists(cmdHistoryPath)
	if err != nil {
		rootCmd.Printf("Error: Can not read command history file on %s, may be missing permissions:\n%s\n", cmdHistoryPath, err)
		return
	}
	if !exists {
		if err := file.CreateMissingDirsAndFileWithRWPerms(cmdHistoryPath, []byte{}); err != nil {
			rootCmd.Printf("Error: Can not create command history file on %s, may be missing permissions:\n%s\n", cmdHistoryPath, err)
		}
	}
	var p = &cobraprompt.CobraPrompt{
		ShowHelpCommandAndFlags:  true,
		ShowHiddenFlags:          true,
		SuggestFlagsWithoutDash:  true,
		DisableCompletionCommand: true,
		AddDefaultExitCommand:    true,
		GoPromptOptions: []goprompt.Option{
			goprompt.OptionTitle("Hazelcast Client"),
			goprompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
				return fmt.Sprintf("hzc %s@%s> ", config.GetClusterAddress(cnfg), cnfg.Cluster.Name), true
			}),
			goprompt.OptionMaxSuggestion(10),
			goprompt.OptionCompletionOnDown(),
		},
		OnErrorFunc: func(err error) {
			errStr := HandleError(err)
			rootCmd.Println(errStr)
			return
		},
	}
	rootCmd.Println("Connecting to the cluster ...")
	if _, err := internal.ConnectToCluster(ctx, cnfg); err != nil {
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
	history, err := ReadCmdHistory(cmdHistoryPath)
	if err != nil {
		rootCmd.Printf("Error: Something went wrong reading command history file on %s:\n%s\n", cmdHistoryPath, err)
		return
	}
	p.Run(ctx, rootCmd, cnfg, history)
	if err := saveCommandHistory(cmdHistoryPath, history.Histories, 100); err != nil {
		rootCmd.Printf("Error: Can not save command history to %s:\n%s\n", cmdHistoryPath, err)
	}
	return
}

func saveCommandHistory(cmdHistoryPath string, commands []string, numberOfCommandsToSave int) error {
	if len(commands) > numberOfCommandsToSave {
		commands = commands[len(commands)-numberOfCommandsToSave:]
	}
	// automatically truncates the file, removing older commands
	f, err := os.Create(cmdHistoryPath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, c := range commands {
		if _, err := f.WriteString(fmt.Sprintln(c)); err != nil {
			return err
		}
	}
	return nil
}

func ReadCmdHistory(path string) (*goprompt.History, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	history := goprompt.NewHistory()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		history.Add(scanner.Text())
	}
	return history, scanner.Err()
}

func updateConfigWithFlags(rootCmd *cobra.Command, cnfg *config.Config, programArgs []string, globalFlagValues *config.GlobalFlagValues) error {
	// parse global persistent flags
	subCmd, flags, _ := rootCmd.Find(programArgs)
	// fall back to cmd.Help, even if there is error
	_ = subCmd.ParseFlags(flags)
	// initialize config from file
	err := config.ReadAndMergeWithFlags(globalFlagValues, cnfg)
	return err
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
