//go:build base

package config

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Cmd struct{}

func (cm Cmd) Init(cc plug.InitContext) error {
	cc.SetTopLevel(true)
	cc.SetCommandUsage("config [command]")
	help := "Show, add or change configuration"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm Cmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("config", &Cmd{}))
}
