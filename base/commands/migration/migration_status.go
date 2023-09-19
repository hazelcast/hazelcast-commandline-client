//go:build std || migration

package migration

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type StatusCmd struct{}

func (s StatusCmd) Unwrappable() {}

func (s StatusCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("status")
	cc.SetCommandGroup("migration")
	help := "Get status of the data migration in progress"
	cc.SetCommandHelp(help, help)
	return nil
}

func (s StatusCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary(banner)
	sts := NewStatusStages(ec.Logger())
	sp := stage.NewFixedProvider(sts.Build(ctx, ec)...)
	if _, err := stage.Execute(ctx, ec, any(nil), sp); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK")
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("status", &StatusCmd{}))
}
