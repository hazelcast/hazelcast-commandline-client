//go:build migration

package migration

import (
	"context"
	"errors"
	"time"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
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
	cc.SetPositionalArgCount(1, 1)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "start the migration without confirmation")
	return nil
}

func (StartCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	configDir := ec.Args()[0]
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary(`Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.
	
Selected data structures in the source cluster will be migrated to the target cluster.	
`)
	if !ec.Props().GetBool(clc.FlagAutoYes) {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Proceed?")
		if err != nil {
			return clcerrors.ErrUserCancelled
		}
		if !yes {
			return clcerrors.ErrUserCancelled
		}
	}
	var ci *hazelcast.ClientInternal
	var q *hazelcast.Queue
	var err error
	ec.PrintlnUnnecessary("")
	stages := []stage.Stage{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func: func(status stage.Statuser) error {
				ci, err = ec.ClientInternal(ctx)
				if err != nil {
					return err
				}
				q, err = ci.Client().GetQueue(ctx, startQueueName)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			ProgressMsg: "Starting the migration",
			SuccessMsg:  "Started the migration",
			FailureMsg:  "Could not start the migration",
			Func: func(status stage.Statuser) error {
				bundle, err := bundleDirAsJSON(configDir)
				if err != nil {
					return err
				}
				if err := q.Put(ctx, bundle); err != nil {
					return err
				}
				return nil
			},
		},
		{
			ProgressMsg: "Migrating the cluster",
			SuccessMsg:  "Migrated the cluster",
			FailureMsg:  "Could not migrate the cluster",
			Func: func(status stage.Statuser) error {
				m, err := ci.Client().GetMap(ctx, statusMapName)
				if err != nil {
					return err
				}
				for {
					s, err := m.Get(ctx, statusMapEntryName)
					if err != nil {
						return err
					}
					switch s {
					case statusComplete:
						return nil
					case statusCanceled:
						return clcerrors.ErrUserCancelled
					case statusFailed:
						return errors.New("migration failed")
					}
					time.Sleep(5 * time.Second)
				}
			},
		},
	}
	sp := stage.NewFixedProvider(stages...)
	if err := stage.Execute(ctx, ec, sp); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK Migration completed successfully.")
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("start", &StartCmd{}))
}
