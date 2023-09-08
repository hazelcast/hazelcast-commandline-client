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

type CancelStages struct {
	ci                       *hazelcast.ClientInternal
	cancelQueue              *hazelcast.Queue
	migrationsInProgressList *hazelcast.List
	migrationID              string
}

func NewCancelStages() *CancelStages {
	return &CancelStages{}
}

func (st *CancelStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage {
	return []stage.Stage{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ctx, ec),
		},
		{
			ProgressMsg: "Canceling the migration",
			SuccessMsg:  "Canceled the migration",
			FailureMsg:  "Could not cancel the migration",
			Func:        st.cancelStage(ctx),
		},
	}
}

func (st *CancelStages) connectStage(ctx context.Context, ec plug.ExecContext) func(stage.Statuser) error {
	return func(status stage.Statuser) error {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return err
		}
		st.migrationsInProgressList, err = st.ci.Client().GetList(ctx, MigrationsInProgressList)
		if err != nil {
			return err
		}
		all, err := st.migrationsInProgressList.GetAll(ctx)
		if err != nil {
			return err
		}
		if len(all) == 0 {
			return fmt.Errorf("there are no migrations are in progress on migration cluster")
		}
		var mip MigrationInProgress
		m := all[0].(serialization.JSON)
		err = json.Unmarshal(m, &mip)
		if err != nil {
			return fmt.Errorf("parsing migration in progress: %w", err)
		}
		st.migrationID = mip.MigrationID
		st.cancelQueue, err = st.ci.Client().GetQueue(ctx, CancelQueue)
		return err
	}
}

func (st *CancelStages) cancelStage(ctx context.Context) func(stage.Statuser) error {
	return func(statuser stage.Statuser) error {
		c := CancelItem{ID: st.migrationID}
		b, err := json.Marshal(c)
		if err != nil {
			return err
		}
		st.cancelQueue.Put(ctx, serialization.JSON(b))
		return err
	}
}

type MigrationInProgress struct {
	MigrationID string `json:"migrationId"`
}

type CancelItem struct {
	ID string `json:"id"`
}
