package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type HomeCommand struct{}

func (hc HomeCommand) Init(cc plug.InitContext) error {
	help := "Print the CLC home directory, optionally by joining the given sub-path"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 1)
	cc.SetCommandUsage("home [sub-path] [flags]")
	return nil
}

func (hc HomeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	dir := paths.Home()
	args := ec.Args()
	if len(args) > 0 {
		dir = filepath.Join(append([]string{dir}, args...)...)
	}
	I2(fmt.Fprintln(ec.Stdout(), dir))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("home", &HomeCommand{}))
}
