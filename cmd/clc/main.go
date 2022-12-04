package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
)

func bye(err error) {
	_, _ = fmt.Fprintf(os.Stderr, err.Error())
	os.Exit(1)
}

func main() {
	// do not exit prematurely on Windows
	cobra.MousetrapHelpText = ""
	// disable color mode
	color.NoColor = true
	args := os.Args[1:]
	cfgPath, logPath, logLevel, err := clc.ExtractStartupArgs(args)
	if err != nil {
		bye(err)
	}
	_, arg0 := filepath.Split(os.Args[0])
	m, err := clc.NewMain(arg0, cfgPath, logPath, logLevel, os.Stdout, os.Stderr)
	if err != nil {
		bye(err)
	}
	err = m.Execute(args)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// ignoring the error here
	_ = m.Exit()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
