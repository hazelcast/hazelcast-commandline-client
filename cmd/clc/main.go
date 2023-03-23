package main

import (
	"fmt"
	"os"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
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
	hazelcast.SetDefaultCompactDeserializer(serialization.NewGenericCompactDeserializer())
	hazelcast.SetDefaultPortableDeserializer(serialization.NewGenericPortableSerializer())
	m, err := clc.NewMain("clc", cfgPath, logPath, logLevel, os.Stdout, os.Stderr)
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
