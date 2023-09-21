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
	ci                       *hazelcast.ClientInternal
	migrationsInProgressList *hazelcast.List
	statusMap                *hazelcast.Map
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
	MigrationID string `json:"migrationId"`
}

func (st *StatusStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		st.migrationsInProgressList, err = st.ci.Client().GetList(ctx, MigrationsInProgressList)
		if err != nil {
			return nil, err
		}
		all, err := st.migrationsInProgressList.GetAll(ctx)
		if err != nil {
			return nil, err
		}
		if len(all) == 0 {
			return nil, fmt.Errorf("there are no migrations are in progress on migration cluster")
		}
		var mip MigrationInProgress
		m := all[0].(serialization.JSON)
		err = json.Unmarshal(m, &mip)
		if err != nil {
			return nil, fmt.Errorf("parsing migration in progress: %w", err)
		}
		st.statusMap, err = st.ci.Client().GetMap(ctx, StatusMapName)
		if err != nil {
			return nil, err
		}
		return mip.MigrationID, err
	}
}
