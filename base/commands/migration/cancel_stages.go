//go:build migration

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

func (st *CancelStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ec),
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
		st.migrationID = mip.MigrationID
		st.cancelQueue, err = st.ci.Client().GetQueue(ctx, CancelQueue)
		return nil, err
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
		return nil, waitForCancel(ctx, st.ci, st.migrationID)
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

func fetchMigrationStatus(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.status') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	res, err := ci.Client().SQL().Execute(ctx, q)
	if err != nil {
		return "", err
	}
	it, err := res.Iterator()
	if err != nil {
		return "", err
	}
	if it.HasNext() { // single iteration is enough that we are reading single result for a single migration
		row, err := it.Next()
		if err != nil {
			return "", err
		}
		r, err := row.Get(0)
		var m string
		if err = json.Unmarshal(r.(serialization.JSON), &m); err != nil {
			return "", err
		}
		return m, nil
	}
	return "", nil
}

type MigrationInProgress struct {
	MigrationID string `json:"migrationId"`
}

type CancelItem struct {
	ID string `json:"id"`
}
