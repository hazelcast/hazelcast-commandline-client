//go:build std || migration

package migration

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

type StatusStages struct {
	ci *hazelcast.ClientInternal
}

func NewStatusStages() *StatusStages {
	return &StatusStages{}
}

func (st *StatusStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ec),
		},
		{
			ProgressMsg: "Finding migration/estimation in progress",
			SuccessMsg:  "Found migration/estimation in progress",
			FailureMsg:  "Could not find a migration/estimation in progress",
			Func:        st.findMigrationInProgress(ec),
		},
	}
}

func (st *StatusStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (st *StatusStages) findMigrationInProgress(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		m, err := findMigrationInProgress(ctx, st.ci)
		if err != nil {
			return nil, err
		}
		return m.MigrationID, err
	}
}
