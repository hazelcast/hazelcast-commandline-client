//go:build migration

package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
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
	ec.PrintlnUnnecessary(`Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.

Task Summary
============

Data Structures to be Migrated:
	- 128 IMaps
	- 32 Replicated Maps
	
Source Cluster:
	ID: c193bea4-1bfb-419a-81b7-4bd34c3c17d5
	Name: prod123
	Version: 5.2.3
	Members:
		- 100.200.300.400
		- 100.200.300.401
		- 200.200.300.402
	
Target Cluster (Viridian):
	ID: 6e71c793-e5cb-4f03-8d51-d85bd396f5f8
	Name: pr-d993dw
	Version: 5.3.1
	
	
Selected data structures in the source cluster will be migrated to the target cluster.	
Once started, the migration will continue even if this application is terminated.`)
	p := prompt.New(ec.Stdin(), ec.Stdout())
	yes, err := p.YesNo("Proceed?")
	if err != nil {
		return errors.ErrUserCancelled
	}
	if !yes {
		return errors.ErrUserCancelled
	}
	ec.PrintlnUnnecessary("")
	remaining := 5 * time.Minute
	var stages []stage.Stage
	stages = append(stages, stage.Stage{
		ProgressMsg: "Validating the setup",
		SuccessMsg:  "Validated the setup",
		FailureMsg:  "Could not validate the setup",
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
		FailureMsg:  "Could not connect to the source cluster",
		Func: func(status stage.Statuser) error {
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	stages = append(stages, stage.Stage{
		ProgressMsg: "Connecting to the target cluster",
		SuccessMsg:  "Connected to the target cluster",
		FailureMsg:  "Could not connect to the target cluster",
		Func: func(status stage.Statuser) error {
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	for i := 0; i < 128; i++ {
		name := fmt.Sprintf("map-%04d", i)
		stages = append(stages, stage.Stage{
			ProgressMsg: fmt.Sprintf("Migrating IMap: %s", name),
			SuccessMsg:  fmt.Sprintf("Migrated IMap: %s", name),
			FailureMsg:  fmt.Sprintf("Could not migrate IMap: %s", name),
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
	for i := 0; i < 128; i++ {
		name := fmt.Sprintf("repmap-%04d", i)
		stages = append(stages, stage.Stage{
			ProgressMsg: fmt.Sprintf("Migrating Replicated Map: %s", name),
			SuccessMsg:  fmt.Sprintf("Migrated Replicated Map: %s", name),
			FailureMsg:  fmt.Sprintf("Could not migrate Replicated Map: %s", name),
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
