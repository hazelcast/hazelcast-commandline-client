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

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/commands"
	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	rootCmd, persistentFlags := commands.NewRoot()
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetOut(os.Stdout)
	// parse global persistent flags
	subcmd, flags, err := rootCmd.Find(os.Args[1:])
	err = subcmd.ParseFlags(flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitError)
	}
	// initialize config from file
	conf, err := config.Get(*persistentFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitError)
	}
	ctx := config.ToContext(context.Background(), conf)
	isInteractive := IsInteractiveCall(rootCmd, os.Args[1:])
	if isInteractive {
		RunCmdInteractively(ctx, rootCmd, conf)
		os.Exit(exitOK)
	}
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		HandleError(rootCmd, err)
		os.Exit(exitError)
	}
	os.Exit(exitOK)
}

func HandleError(cmd *cobra.Command, err error) {
	errStr := fmt.Sprintf("Unknown Error: %s\n", err.Error())
	var loggable hzcerror.LoggableError
	if errors.As(err, &loggable) {
		errStr = fmt.Sprintf("Error: %s\n", loggable.VerboseError())
	}
	cmd.PrintErrln(errStr)
}
