package sampleplug

import (
	"fmt"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct {
}

func (c Command) Init(ctx plug.CommandContext) error {
	return nil
}

func (c Command) Exec(ctx plug.ExecContext) error {
	I2(fmt.Fprintf(ctx.Stdout(), "Hello, World!\n"))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("util:hello", &Command{}))
}
