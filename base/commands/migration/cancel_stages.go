//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type CancelStages struct {
	ci          *hazelcast.ClientInternal
	cancelQueue *hazelcast.Queue
	migrationID string
}

func NewCancelStages() *CancelStages {
	return &CancelStages{}
}

func (st *CancelStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ec),
		},
		{
			ProgressMsg: "Finding migration in progress",
			SuccessMsg:  "Found migration in progress",
			FailureMsg:  "Could not find a migration in progress",
			Func:        st.findMigrationInProgress(ec),
		},
		{
			ProgressMsg: "Canceling the migration",
			SuccessMsg:  "Canceled the migration",
			FailureMsg:  "Could not cancel the migration",
			Func:        st.cancelStage(),
		},
	}
}

func (st *CancelStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		st.cancelQueue, err = st.ci.Client().GetQueue(ctx, CancelQueue)
		return nil, nil
	}
}

func (st *CancelStages) findMigrationInProgress(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		m, err := findMigrationInProgress(ctx, st.ci)
		if err != nil {
			return nil, err
		}
		st.migrationID = m.MigrationID
		return nil, nil
	}
}

func (st *CancelStages) cancelStage() func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		c := CancelItem{ID: st.migrationID}
		b, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		err = st.cancelQueue.Put(ctx, serialization.JSON(b))
		if err != nil {
			return nil, err
		}
		childCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return nil, waitForCancel(childCtx, st.ci, st.migrationID)
	}
}

func waitForCancel(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) error {
	for {
		s, err := fetchMigrationStatus(ctx, ci, migrationID)
		if err != nil {
			return err
		}
		if Status(s) == StatusCanceling {
			return nil
		}
	}
}

type CancelItem struct {
	ID string `json:"id"`
}
