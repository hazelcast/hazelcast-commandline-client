//go:build base

package config

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ImportCmd struct{}

func (cm ImportCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("import TARGET SOURCE")
	help := "Imports configuration from an arbitrary source"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(2, 2)
	return nil
}

func (cm ImportCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	target := ec.Args()[0]
	src := ec.Args()[1]
	path, err := config.ImportSource(ctx, ec, target, src)
	if err != nil {
		return err
	}
	if ec.Interactive() || ec.Props().GetBool(clc.PropertyVerbose) {
		I2(fmt.Fprintf(ec.Stdout(), "Created configuration at: %s\n", path))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("config:import", &ImportCmd{}))
}
