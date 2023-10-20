//go:build migration

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
	cc.AddStringFlag(flagOutputDir, "o", "", false, "output directory for the migration report, if not given current directory is used")
	cc.SetCommandHelp(help, help)
	return nil
}

func (s StatusCmd) Exec(ctx context.Context, ec plug.ExecContext) (err error) {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary(banner)
	sts := NewStatusStages()
	sp := stage.NewFixedProvider(sts.Build(ctx, ec)...)
	mID, err := stage.Execute(ctx, ec, any(nil), sp)
	if err != nil {
		return err
	}
	defer func() {
		finalizeErr := finalizeMigration(ctx, ec, ci, mID.(string), ec.Props().GetString(flagOutputDir))
		if err == nil {
			err = finalizeErr
		}
	}()
	mStages, err := createMigrationStages(ctx, ec, ci, mID.(string))
	if err != nil {
		return err
	}
	mp := stage.NewFixedProvider(mStages...)
	_, err = stage.Execute(ctx, ec, any(nil), mp)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK")
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("status", &StatusCmd{}))
}
