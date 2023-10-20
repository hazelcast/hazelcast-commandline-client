//go:build std || migration

package migration

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type CancelCmd struct{}

func (c CancelCmd) Unwrappable() {}

func (c CancelCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("cancel")
	cc.SetCommandGroup("migration")
	help := "Cancel the data migration"
	cc.SetCommandHelp(help, help)
	return nil
}

func (c CancelCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	sts := NewCancelStages()
	sp := stage.NewFixedProvider(sts.Build(ctx, ec)...)
	if _, err := stage.Execute(ctx, ec, any(nil), sp); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK Migration canceled successfully.")
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("cancel", &CancelCmd{}))
}
