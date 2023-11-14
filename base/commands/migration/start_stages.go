//go:build std || migration

package migration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

type StartStages struct {
	migrationID string
	configDir   string
	ci          *hazelcast.ClientInternal
	startQueue  *hazelcast.Queue
	logger      log.Logger
}

func NewStartStages(logger log.Logger, migrationID, configDir string) (*StartStages, error) {
	if migrationID == "" {
		return nil, errors.New("migrationID is required")
	}
	return &StartStages{
		migrationID: migrationID,
		configDir:   configDir,
		logger:      logger,
	}, nil
}

func (st *StartStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ec),
		},
		{
			ProgressMsg: "Starting the migration",
			SuccessMsg:  fmt.Sprintf("Started the migration with ID: %s", st.migrationID),
			FailureMsg:  "Could not start the migration",
			Func:        st.startStage(),
		},
		{
			ProgressMsg: "Doing pre-checks",
			SuccessMsg:  "Pre-checks complete",
			FailureMsg:  "Could not complete pre-checks",
			Func:        st.preCheckStage(),
		},
	}
}

func (st *StartStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		st.startQueue, err = st.ci.Client().GetQueue(ctx, StartQueueName)
		if err != nil {
			return nil, fmt.Errorf("retrieving the start Queue: %w", err)
		}
		return nil, nil
	}
}

func (st *StartStages) startStage() func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		cb, err := makeConfigBundle(st.configDir, st.migrationID)
		if err != nil {
			return nil, fmt.Errorf("making configuration bundle: %w", err)
		}
		if err = st.startQueue.Put(ctx, cb); err != nil {
			return nil, fmt.Errorf("updating start Queue: %w", err)
		}
		childCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		if err = waitForMigrationToStart(childCtx, st.ci, st.migrationID); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (st *StartStages) preCheckStage() func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		childCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		if err := WaitForMigrationToBeInProgress(childCtx, st.ci, st.migrationID); err != nil {
			return nil, fmt.Errorf("waiting for prechecks to complete: %w", err)
		}
		return nil, nil
	}
}

func waitForMigrationToStart(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		status, err := fetchMigrationStatus(ctx, ci, migrationID)
		if err != nil {
			if errors.Is(err, migrationStatusNotFoundErr) {
				// migration status will not be available for a while, so we should wait for it
				continue
			}
			return err
		}
		if Status(status) == StatusFailed {
			errs, err := fetchMigrationErrors(ctx, ci, migrationID)
			if err != nil {
				return fmt.Errorf("migration failed and dmt cannot fetch migration errors: %w", err)
			}
			return errors.New(errs)
		}
		if Status(status) == StatusStarted || Status(status) == StatusInProgress || Status(status) == StatusComplete {
			return nil
		}
	}
}
