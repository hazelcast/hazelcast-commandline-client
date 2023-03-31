package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc"
	cmd "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
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
	cp, err := config.NewFileProvider(cfgPath)
	if err != nil {
		bye(err)
	}
	m, err := cmd.NewMain("clc", cfgPath, cp, logPath, logLevel, clc.StdIO())
	if err != nil {
		bye(err)
	}
	err = m.Execute(args...)
	if err != nil {
		// print the error only if it wasn't printed before
		if _, ok := err.(errors.WrappedError); !ok {
			fmt.Println("Error:", err)
		}
	}
	// ignoring the error here
	_ = m.Exit()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
