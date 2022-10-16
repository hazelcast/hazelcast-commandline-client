package main

import (
	"fmt"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func main() {
	m := clc.NewMain(false)
	if err := m.Execute(); err != nil {
		I2(fmt.Fprintln(os.Stderr, err.Error()))
	}
	m.Exit()
}
