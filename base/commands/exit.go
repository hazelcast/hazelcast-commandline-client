//go:build base

package commands

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	puberrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ExitCommand struct{}

func (ex ExitCommand) Init(cc plug.InitContext) error {
	help := "Exit with the given status code"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, math.MaxInt)
	cc.SetCommandUsage("exit [STATUS CODE] [flags]")
	return nil
}

func (ex ExitCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	args := ec.Args()
	if len(args) > 0 {
		code, err := strconv.Atoi(args[0])
		if err != nil || code < 0 || code > 255 {
			return fmt.Errorf("Given status code should range between 0 and 255.")
		}
		return puberrors.NewExitError(code)
	}
	return fmt.Errorf("Usage: %sexit [STATUS CODE] [flags]", shell.CmdPrefix)

}

func init() {
	Must(plug.Registry.RegisterCommand("exit", &ExitCommand{}))
}
