//go:build base

package commands

import (
	"context"
	"fmt"
	"math"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type EchoCommand struct{}

func (ecc EchoCommand) Init(cc plug.InitContext) error {
	help := "Prints the given string in scripting mode"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, math.MaxInt)
	cc.SetCommandUsage("echo [STRING] [flags]")
	return nil
}

func (ecc EchoCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	args := ec.Args()
	if len(args) > 0 {
		I2(fmt.Fprintln(ec.Stdout(), args[0]))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("echo", &EchoCommand{}))
}
