//go:build std || config

package config

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("config")
	cc.SetTopLevel(true)
	help := "Show, add or change configuration"
	cc.SetCommandHelp(help, help)
	return nil
}

func (Command) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("config", &Command{}))
}
