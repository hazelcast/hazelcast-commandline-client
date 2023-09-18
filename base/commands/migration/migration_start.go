//go:build std || migration

package migration

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-go-client"
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
	cc.AddStringFlag(flagOutputDir, "o", "", false, "output directory for the migration report, if not given current directory is used")
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
			return clcerrors.ErrUserCancelled
		}
		if !yes {
			return clcerrors.ErrUserCancelled
		}
	}
	ec.PrintlnUnnecessary("")
	var updateTopic *hazelcast.Topic
	sts := NewStartStages(updateTopic, MakeMigrationID(), ec.Args()[0], ec.Props().GetString(flagOutputDir))
	if !sts.topicListenerID.Default() && sts.updateTopic != nil {
		if err := sts.updateTopic.RemoveListener(ctx, sts.topicListenerID); err != nil {
			return err
		}
	}
	sp := stage.NewFixedProvider(sts.Build(ctx, ec)...)
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
