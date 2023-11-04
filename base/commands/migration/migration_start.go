//go:build std || migration

package migration

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type StartCmd struct{}

func (StartCmd) Unwrappable() {}

func (StartCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("start")
	cc.SetCommandGroup("migration")
	help := "Start the data migration"
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "start the migration without confirmation")
	cc.AddStringArg(argDMTConfig, argTitleDMTConfig)
	cc.AddStringFlag(flagOutputDir, "o", "", false, "output directory for the migration report, if not given current directory is used")
	return nil
}

func (StartCmd) Exec(ctx context.Context, ec plug.ExecContext) (err error) {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary(`Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.
	
Selected data structures in the source cluster will be migrated to the target cluster.

Prerequisites:

	* Only a single migration can be run simultaneously.
	* The target cluster must be an enterprise cluster or Viridian cluster.
	* At least one data structure name must be provided in the migration configuration file.
	* The data structures given in the migration configuration must exist in the source cluster.

Further Information:

	* Existing data structures on the target cluster will be skipped.
	* It is recommended that the source cluster is put in read-only mode before the migration starts.
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
	ec.PrintlnUnnecessary("")
	mID := MakeMigrationID()
	defer func() {
		finalizeErr := finalizeMigration(ctx, ec, ci, mID, ec.Props().GetString(flagOutputDir))
		if err == nil {
			err = finalizeErr
		}
	}()
	sts, err := NewStartStages(ec.Logger(), mID, ec.GetStringArg(argDMTConfig))
	if err != nil {
		return err
	}
	sp := stage.NewFixedProvider(sts.Build(ctx, ec)...)
	if _, err = stage.Execute(ctx, ec, any(nil), sp); err != nil {
		return err
	}
	mStages, err := createMigrationStages(ctx, ec, ci, mID)
	if err != nil {
		return err
	}
	mp := stage.NewFixedProvider(mStages...)
	_, err = stage.Execute(ctx, ec, any(nil), mp)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK Migration completed successfully.")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Migration ID",
			Type:  serialization.TypeString,
			Value: mID,
		},
	})
}

func init() {
	check.Must(plug.Registry.RegisterCommand("start", &StartCmd{}))
}
