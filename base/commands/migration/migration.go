//go:build migration

package migration

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Cmd struct{}

func (cm Cmd) Init(cc plug.InitContext) error {
	cc.SetTopLevel(true)
	cc.SetCommandUsage("migration [command]")
	help := "Data migration operations"
	cc.SetCommandHelp(help, help)
	cc.AddCommandGroup("migration", "Data Migration")
	cc.SetCommandGroup("migration")
	return nil
}

func (cm Cmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("migration", &Cmd{}))
}
