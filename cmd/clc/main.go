package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc"
	cmd "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	hzerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

const (
	ExitCodeSuccess        = 0
	ExitCodeGenericFailure = 1
	ExitCodeTimeout        = 2
)

func bye(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(1)
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
	m, err := cmd.NewMain(name, cfgPath, cp, logPath, logLevel, clc.StdIO())
	if err != nil {
		bye(err)
	}
	err = m.Execute(context.Background(), args...)
	if err != nil {
		// print the error only if it wasn't printed before
		if _, ok := err.(hzerrors.WrappedError); !ok {
			fmt.Println(cmd.MakeErrStr(err))
		}
	}
	// ignoring the error here
	_ = m.Exit()
	if err != nil {
		// keeping the hzerrors.ErrTimeout for now
		// it may be useful to send that error in the future. --YT
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, hzerrors.ErrTimeout) {
			os.Exit(ExitCodeTimeout)
		}
		os.Exit(ExitCodeGenericFailure)
	}
	os.Exit(ExitCodeSuccess)
}
