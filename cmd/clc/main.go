package main

import (
	"os"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
)

func main() {
	m := clc.NewMain(false)
	// ingoring the error here
	err := m.Execute(os.Args[1:])
	// ingoring the error here
	_ = m.Exit()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
