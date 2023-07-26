//go:build migration

package migration

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
)

type StartCmd struct{}

func (StartCmd) Unwrappable() {}

func (StartCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("start [dmt-config] [flags]")
	cc.SetCommandGroup("migration")
	help := "Start the data migration"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 1)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "start the migration without confirmation")
	return nil
}

func (StartCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary(`Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.
	
Selected data structures in the source cluster will be migrated to the target cluster.	
`)
	if !ec.Props().GetBool(clc.FlagAutoYes) {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Proceed?")
		if err != nil {
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	ec.PrintlnUnnecessary("")
	stages := []stage.Stage{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func: func(status stage.Statuser) error {
				time.Sleep(1 * time.Second)
				return nil
			},
		},
		{
			ProgressMsg: "Starting the migration",
			SuccessMsg:  "Started the migration",
			FailureMsg:  "Could not start the migration",
			Func: func(status stage.Statuser) error {
				time.Sleep(1 * time.Second)
				return nil
			},
		},
		{
			ProgressMsg: "Migrating the cluster",
			SuccessMsg:  "Migrated the cluster",
			FailureMsg:  "Could not migrate the cluster",
			Func: func(status stage.Statuser) error {
				time.Sleep(10 * time.Second)
				return nil
			},
		},
	}
	sp := stage.NewFixedProvider(stages...)
	if err := stage.Execute(ctx, ec, sp); err != nil {
		ec.PrintlnUnnecessary("FAIL Migration failed: " + err.Error())
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK Migration completed successfully.")
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("start", &StartCmd{}))
}
