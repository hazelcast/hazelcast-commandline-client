package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
)

func bye(err error) {
	_, _ = fmt.Fprintf(os.Stderr, err.Error())
	os.Exit(1)
}

func main() {
	// disable color mode
	color.NoColor = true
	args := os.Args[1:]
	cfgPath, logPath, logLevel, err := clc.ExtractStartupArgs(args)
	if err != nil {
		bye(err)
	}
	m, err := clc.NewMain(cfgPath, logPath, logLevel, os.Stdout, os.Stderr)
	if err != nil {
		bye(err)
	}
	err = m.Execute(args)
	// ignoring the error here
	_ = m.Exit()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
