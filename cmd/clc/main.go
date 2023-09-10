package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc"
	cmd "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	hzerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

const (
	ExitCodeSuccess        = 0
	ExitCodeGenericFailure = 1
	ExitCodeTimeout        = 2
	ExitCodeUserCanceled   = 3
)

func bye(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(ExitCodeGenericFailure)
}

func main() {
	// do not exit prematurely on Windows
	cobra.MousetrapHelpText = ""
	args := os.Args[1:]
	cfgPath, logPath, logLevel, err := cmd.ExtractStartupArgs(args)
	if err != nil {
		bye(err)
	}
	cp, err := config.NewWizardProvider(cfgPath)
	if err != nil {
		bye(err)
	}
	_, name := filepath.Split(os.Args[0])
	stdio := clc.StdIO()
	m, err := cmd.NewMain(name, cfgPath, cp, logPath, logLevel, stdio)
	if err != nil {
		bye(err)
	}
	err = m.Execute(context.Background(), args...)
	if err != nil {
		// print the error only if it wasn't printed before
		if _, ok := err.(hzerrors.WrappedError); !ok {
			if !hzerrors.IsUserCancelled(err) {
				check.I2(fmt.Fprintln(stdio.Stderr, hzerrors.MakeString(err)))
			}
		}
	}
	// ignoring the error here
	_ = m.Exit()
	if err != nil {
		if hzerrors.IsTimeout(err) {
			os.Exit(ExitCodeTimeout)
		}
		if hzerrors.IsUserCancelled(err) {
			os.Exit(ExitCodeUserCanceled)
		}
		os.Exit(ExitCodeGenericFailure)
	}
	os.Exit(ExitCodeSuccess)
}
