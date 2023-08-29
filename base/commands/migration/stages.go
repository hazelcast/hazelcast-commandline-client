package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Stages struct {
	migrationID string
	configDir   string
	ci          *hazelcast.ClientInternal
	startQueue  *hazelcast.Queue
	statusMap   *hazelcast.Map
}

func NewStages(migrationID, configDir string) *Stages {
	if migrationID == "" {
		panic("migrationID is required")
	}
	return &Stages{
		migrationID: migrationID,
		configDir:   configDir,
	}
}

func (st *Stages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage {
	return []stage.Stage{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ctx, ec),
		},
		{
			ProgressMsg: "Starting the migration",
			SuccessMsg:  "Started the migration",
			FailureMsg:  "Could not start the migration",
			Func:        st.startStage(ctx),
		},
		{
			ProgressMsg: "Migrating the cluster",
			SuccessMsg:  "Migrated the cluster",
			FailureMsg:  "Could not migrate the cluster",
			Func:        st.migrateStage(ctx),
		},
	}
}

func (st *Stages) connectStage(ctx context.Context, ec plug.ExecContext) func(stage.Statuser) error {
	return func(status stage.Statuser) error {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return err
		}
		st.startQueue, err = st.ci.Client().GetQueue(ctx, startQueueName)
		if err != nil {
			return err
		}
		st.statusMap, err = st.ci.Client().GetMap(ctx, makeStatusMapName(st.migrationID))
		if err != nil {
			return err
		}
		return nil
	}
}

func (st *Stages) startStage(ctx context.Context) func(stage.Statuser) error {
	return func(status stage.Statuser) error {
		if err := st.statusMap.Delete(ctx, statusMapEntryName); err != nil {
			return err
		}
		var cb configBundle
		cb.MigrationID = st.migrationID
		if err := cb.Walk(st.configDir); err != nil {
			return err
		}
		b, err := json.Marshal(cb)
		if err != nil {
			return err
		}
		if err := st.startQueue.Put(ctx, serialization.JSON(b)); err != nil {
			return err
		}
		if err = st.waitStatusResult(ctx, 10, 1*time.Second); err != nil {
			return err
		}
		return nil
	}
}

func (st *Stages) migrateStage(ctx context.Context) func(statuser stage.Statuser) error {
	return func(status stage.Statuser) error {
		return st.waitStatusResult(ctx, -1, 5*time.Second)
	}
}

func (st *Stages) waitStatusResult(ctx context.Context, maxRetries int, period time.Duration) error {
	tryInfinitely := maxRetries < 0
	tryCount := 0
	for {
		if !tryInfinitely {
			if maxRetries == tryCount {
				break
			}
			tryCount += 1
		}
		s, err := st.readStatus(ctx)
		if err != nil {
			return fmt.Errorf("reading status: %w", err)
		}
		switch s {
		case statusInProgress, statusComplete:
			return nil
		case statusCanceled:
			return clcerrors.ErrUserCancelled
		case statusFailed:
			return errors.New("migration failed")
		}
		time.Sleep(period)
	}
	return fmt.Errorf("reached timeout while reading status")
}

func (st *Stages) readStatus(ctx context.Context) (status, error) {
	v, err := st.statusMap.Get(ctx, statusMapEntryName)
	if err != nil {
		return statusNone, err
	}
	if v == nil {
		return statusNone, nil
	}
	var b []byte
	if vv, ok := v.(string); ok {
		b = []byte(vv)
	} else if vv, ok := v.(serialization.JSON); ok {
		b = vv
	} else {
		return statusNone, fmt.Errorf("invalid status value")
	}
	var ms migrationStatus
	if err := json.Unmarshal(b, &ms); err != nil {
		return statusNone, fmt.Errorf("unmarshaling status: %w", err)
	}
	return ms.Status, nil
}

func makeStatusMapName(migrationID string) string {
	return "__datamigration_" + migrationID
}

type status string

const (
	statusNone       status = ""
	statusComplete   status = "COMPLETED"
	statusCanceled   status = "CANCELED"
	statusFailed     status = "FAILED"
	statusInProgress status = "IN_PROGRESS"
)

type migrationStatus struct {
	Status status `json:"status"`
}
