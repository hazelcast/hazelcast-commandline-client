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
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/config"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	rootCmd, globalFlagValues := InitRootCmd()
	programArgs := os.Args[1:]
	conf, err := getConfigWithFlags(rootCmd, programArgs, globalFlagValues)
	ExitOnError(err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.ToContext(ctx, conf)
	isInteractive := IsInteractiveCall(rootCmd, programArgs)
	if isInteractive {
		RunCmdInteractively(ctx, rootCmd, conf)
	} else {
		err = RunCmd(ctx, rootCmd)
		ExitOnError(err)
	}
	return
}

func ExitOnError(err error) {
	if err == nil {
		return
	}
	errStr := HandleError(err)
	fmt.Println(errStr)
	os.Exit(exitError)
}
