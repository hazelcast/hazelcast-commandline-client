package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"golang.org/x/exp/slices"

	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
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

func (st *Stages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ec),
		},
		{
			ProgressMsg: "Starting the migration",
			SuccessMsg:  "Started the migration",
			FailureMsg:  "Could not start the migration",
			Func:        st.startStage(),
		},
		{
			ProgressMsg: "Migrating the cluster",
			SuccessMsg:  "Migrated the cluster",
			FailureMsg:  "Could not migrate the cluster",
			Func:        st.migrateStage(),
		},
	}
}

func (st *Stages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		st.startQueue, err = st.ci.Client().GetQueue(ctx, startQueueName)
		if err != nil {
			return nil, err
		}
		st.statusMap, err = st.ci.Client().GetMap(ctx, makeStatusMapName(st.migrationID))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (st *Stages) startStage() func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		if err := st.statusMap.Delete(ctx, statusMapEntryName); err != nil {
			return nil, err
		}
		var cb configBundle
		cb.MigrationID = st.migrationID
		if err := cb.Walk(st.configDir); err != nil {
			return nil, err
		}
		b, err := json.Marshal(cb)
		if err != nil {
			return nil, err
		}
		if err = st.startQueue.Put(ctx, serialization.JSON(b)); err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err = st.waitForStatus(ctx, time.Second, statusInProgress, statusComplete); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (st *Stages) migrateStage() func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		return st.waitForStatus(ctx, 5*time.Second, statusComplete), nil
	}
}

func (st *Stages) waitForStatus(ctx context.Context, waitInterval time.Duration, targetStatuses ...status) error {
	timeoutErr := fmt.Errorf("migration could not be completed: reached timeout while reading status: "+
		"please ensure that you are using Hazelcast's migration cluster distribution and your DMT config points to that cluster: %w",
		context.DeadlineExceeded)
	for {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return timeoutErr
			}
			return fmt.Errorf("migration failed: %w", err)
		}
		s, err := st.readStatus(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return timeoutErr
			}
			return fmt.Errorf("reading status: %w", err)
		}
		switch s {
		case statusComplete:
			return nil
		case statusCanceled:
			return clcerrors.ErrUserCancelled
		case statusFailed:
			return errors.New("migration failed")
		}
		if slices.Contains(targetStatuses, s) {
			return nil
		}
		time.Sleep(waitInterval)
	}
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
