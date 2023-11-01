//go:build std || migration

package migration

import (
	"context"
	"errors"
	"fmt"

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
	statusMap   *hazelcast.Map
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
		st.statusMap, err = st.ci.Client().GetMap(ctx, StatusMapName)
		if err != nil {
			return nil, fmt.Errorf("retrieving the status Map: %w", err)
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
		return nil, nil
	}
}
