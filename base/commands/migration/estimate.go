//go:build std || migration

package migration

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type EstimateCmd struct{}

func (e EstimateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("estimate")
	cc.SetCommandGroup("migration")
	help := "Estimate migration"
	cc.SetCommandHelp(help, help)
	cc.AddStringArg(argDMTConfig, argTitleDMTConfig)
	return nil
}

func (e EstimateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary(fmt.Sprintf(`%s

Estimation usually ends within 15 seconds.`, banner))
	conf := ec.GetStringArg(argDMTConfig)
	if !paths.Exists(conf) {
		return fmt.Errorf("migration config does not exist: %s", conf)
	}
	mID := makeMigrationID()
	stages, err := NewEstimateStages(ec.Logger(), mID, conf)
	if err != nil {
		return err
	}
	sp := stage.NewFixedProvider(stages.Build(ctx, ec)...)
	res, err := stage.Execute(ctx, ec, any(nil), sp)
	if err != nil {
		return err
	}
	resArr := res.([]string)
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary(resArr[0])
	ec.PrintlnUnnecessary(resArr[1])
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK Estimation completed successfully.")
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("estimate", &EstimateCmd{}))
}
