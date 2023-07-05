//go:build migration

package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type StartCmd struct{}

func (cm StartCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("start")
	help := "Start the data migration"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm StartCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.PrintlnUnnecessary("")
	remaining := 5 * time.Minute
	var stages []stage.Stage
	stages = append(stages, stage.Stage{
		ProgressMsg: "Validating the setup",
		SuccessMsg:  "Validated the setup",
		Func: func(status stage.Statuser) error {
			for i := 0; i < 10; i++ {
				time.Sleep(1 * time.Second)
				remaining -= 1 * time.Second
				status.SetRemainingDuration(remaining)
				status.SetProgress(float32(i+1) / 10)
			}
			status.SetProgress(1)
			return nil
		},
	})
	stages = append(stages, stage.Stage{
		ProgressMsg: "Connecting to the source cluster",
		SuccessMsg:  "Connected to the source cluster",
		Func: func(status stage.Statuser) error {
			for i := 0; i < 5; i++ {
				time.Sleep(2 * time.Second)
				remaining -= 2 * time.Second
				status.SetRemainingDuration(remaining)
			}
			return nil
		},
	})
	for i := 0; i < 1000; i++ {
		name := fmt.Sprintf("map.%04d", i)
		stages = append(stages, stage.Stage{
			ProgressMsg: fmt.Sprintf("Migrating IMap: %s", name),
			SuccessMsg:  fmt.Sprintf("Migrated IMap: %s", name),
			FailureMsg:  fmt.Sprintf("Migrating IMap: %s", name),
			Func: func(status stage.Statuser) error {
				for i := 0; i < 3; i++ {
					time.Sleep(1 * time.Second)
					remaining -= 1 * time.Second
					status.SetRemainingDuration(remaining)
					status.SetProgress(float32(i+1) / 3)
				}
				status.SetProgress(1)
				return nil
			},
		})
	}
	stages = append(stages, stage.Stage{
		ProgressMsg: "Cleaning up",
		SuccessMsg:  "Cleaned up",
		FailureMsg:  "Could not clean up",
		Func: func(status stage.Statuser) error {
			for i := 0; i < 5; i++ {
				time.Sleep(2 * time.Second)
				remaining -= 2 * time.Second
				status.SetRemainingDuration(remaining)
			}
			return nil
		},
	})
	sp := stage.NewFixedProvider(stages...)
	return stage.Execute(ctx, ec, sp)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("migration:start", &StartCmd{}))
}
