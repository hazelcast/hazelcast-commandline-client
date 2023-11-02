//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
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
	}
}

type MigrationInProgress struct {
	MigrationID string `json:"id"`
}

func (st *StatusStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		m, err := findMigrationInProgress(ctx, st.ci)
		if err != nil {
			return nil, err
		}
		return m.MigrationID, err
	}
}

func findMigrationInProgress(ctx context.Context, ci *hazelcast.ClientInternal) (MigrationInProgress, error) {
	var mip MigrationInProgress
	q := fmt.Sprintf("SELECT this FROM %s WHERE JSON_VALUE(this, '$.status') IN('STARTED', 'IN_PROGRESS', 'CANCELLING')", StatusMapName)
	r, err := querySingleRow(ctx, ci, q)
	if err != nil {
		return mip, fmt.Errorf("finding migration in progress: %w", err)
	}
	m := r.(serialization.JSON)
	if err = json.Unmarshal(m, &mip); err != nil {
		return mip, fmt.Errorf("parsing migration in progress: %w", err)
	}
	return mip, nil
}
