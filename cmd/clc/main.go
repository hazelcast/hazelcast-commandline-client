package main

import (
	"github.com/hazelcast/hazelcast-commandline-client/clc"
)

func main() {
	m := clc.NewMain(false)
	// ingoring the error here
	_ = m.Execute()
	// ingoring the error here
	_ = m.Exit()
}
