//go:build std || migration

package migration

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type EstimateStages struct {
	migrationID   string
	configDir     string
	ci            *hazelcast.ClientInternal
	estimateQueue *hazelcast.Queue
	statusMap     *hazelcast.Map
	logger        log.Logger
}

func NewEstimateStages(logger log.Logger, migrationID, configDir string) (*EstimateStages, error) {
	if migrationID == "" {
		return nil, errors.New("migrationID is required")
	}
	return &EstimateStages{
		migrationID: migrationID,
		configDir:   configDir,
		logger:      logger,
	}, nil
}

func (es *EstimateStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        es.connectStage(ec),
		},
		{
			ProgressMsg: "Estimating the migration",
			SuccessMsg:  "Estimated the migration",
			FailureMsg:  "Could not estimate the migration",
			Func:        es.estimateStage(),
		},
	}
}

func (es *EstimateStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, s stage.Statuser[any]) (any, error) {
		var err error
		es.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		es.estimateQueue, err = es.ci.Client().GetQueue(ctx, EstimateQueueName)
		if err != nil {
			return nil, fmt.Errorf("retrieving the estimate Queue: %w", err)
		}
		es.statusMap, err = es.ci.Client().GetMap(ctx, StatusMapName)
		if err != nil {
			return nil, fmt.Errorf("retrieving the status Map: %w", err)
		}
		return nil, nil
	}
}

func (es *EstimateStages) estimateStage() func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		cb, err := makeConfigBundle(es.configDir, es.migrationID)
		if err != nil {
			return nil, fmt.Errorf("making configuration bundle: %w", err)
		}
		if err = es.estimateQueue.Put(ctx, cb); err != nil {
			return nil, fmt.Errorf("updating estimate Queue: %w", err)
		}
		childCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err = waitForEstimationToComplete(childCtx, es.ci, es.migrationID, status); err != nil {
			return nil, fmt.Errorf("waiting for estimation to complete: %w", err)
		}
		return fetchEstimationResults(ctx, es.ci, es.migrationID)
	}
}

func waitForEstimationToComplete(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string, stage stage.Statuser[any]) error {
	duration := 16 * time.Second
	interval := 1 * time.Second
	for {
		if duration > 0 {
			stage.SetRemainingDuration(duration)
		} else {
			stage.SetText("Estimation took longer than expected.")
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
				return fmt.Errorf("estimation failed and dmt cannot fetch estimation errors: %w", err)
			}
			return errors.New(errs)
		}
		if Status(status) == StatusComplete {
			return nil
		}
		time.Sleep(interval)
		duration = duration - interval
	}
}

func fetchEstimationResults(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) ([]string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.estimatedTime'), JSON_QUERY(this, '$.estimatedSize') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	res, err := ci.Client().SQL().Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	it, err := res.Iterator()
	if err != nil {
		return nil, err
	}
	if it.HasNext() {
		// single iteration is enough that we are reading single result for a single migration
		row, err := it.Next()
		if err != nil {
			return nil, err
		}
		estimatedTime, err := row.Get(0)
		if err != nil {
			return nil, err
		}
		estimatedSize, err := row.Get(1)
		if err != nil {
			return nil, err
		}
		et, err := MsToSecs(estimatedTime.(serialization.JSON).String())
		if err != nil {
			return nil, err
		}
		es, err := BytesToMegabytes(estimatedSize.(serialization.JSON).String())
		if err != nil {
			return nil, err
		}
		return []string{fmt.Sprintf("Estimated Size: %s ", es), fmt.Sprintf("Estimated Time: %s", et)}, nil
	}
	return nil, errors.New("no rows found")
}

func BytesToMegabytes(bytesStr string) (string, error) {
	bytes, err := strconv.ParseFloat(bytesStr, 64)
	if err != nil {
		return "", err
	}
	mb := bytes / (1024.0 * 1024.0)
	return fmt.Sprintf("%.2f MBs", mb), nil
}

func MsToSecs(ms string) (string, error) {
	milliseconds, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return "", err
	}
	seconds := float64(milliseconds) / 1000.0
	secondsStr := fmt.Sprintf("%.1f sec", seconds)
	return secondsStr, nil
}
