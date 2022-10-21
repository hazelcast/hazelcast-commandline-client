package main

import (
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
)

func main() {
	m := clc.NewMain(false)
	// ingoring the error here
	err := m.Execute()
	// ingoring the error here
	_ = m.Exit()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
