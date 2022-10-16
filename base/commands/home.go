package commands

import (
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type HomeCommand struct{}

func (hc HomeCommand) Init(cc plug.InitContext) error {
	usage := "Print the CLC home directory"
	cc.SetCommandUsage(usage, usage)
	return nil
}

func (hc HomeCommand) Exec(ec plug.ExecContext) error {
	I2(fmt.Fprintln(ec.Stdout(), paths.HomeDir()))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("home", &HomeCommand{}))
}
