package sampleplug

import (
	"context"
	"fmt"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct {
}

func (c Command) Exec(ctx context.Context, ec plug.ExecContext) error {
	I2(fmt.Fprintf(ec.Stdout(), "Hello, World!\n"))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("util:hello", &Command{}))
}
