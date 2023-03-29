package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	puberrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

func bye(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(1)
}

func main() {
	// do not exit prematurely on Windows
	cobra.MousetrapHelpText = ""
	args := os.Args[1:]
	cfgPath, logPath, logLevel, err := clc.ExtractStartupArgs(args)
	if err != nil {
		bye(err)
	}
	m, err := clc.NewMain("clc", cfgPath, logPath, logLevel, os.Stdout, os.Stderr)
	if err != nil {
		bye(err)
	}
	err = m.Execute(args)
	code := -1
	if errors.Unwrap(err) == puberrors.ErrExitWithCode {
		code = err.(*puberrors.ExitWithStatusError).GetCode()
	} else if err != nil {
		fmt.Println("Error:", err)
	}
	// ignoring the error here
	_ = m.Exit()
	if code != -1 {
		os.Exit(code)
	} else if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
