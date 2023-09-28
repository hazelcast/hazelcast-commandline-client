//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	errors2 "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

var timeoutErr = fmt.Errorf("migration could not be completed: reached timeout while reading status: "+
	"please ensure that you are using Hazelcast's migration cluster distribution and your DMT config points to that cluster: %w",
	context.DeadlineExceeded)

func createMigrationStages(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, migrationID string) ([]stage.Stage[any], error) {
	childCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := waitForMigrationToBeCreated(childCtx, ci, migrationID); err != nil {
		return nil, fmt.Errorf("waiting migration to be created: %w", err)
	}
	var stages []stage.Stage[any]
	dss, err := dataStructuresToBeMigrated(ctx, ec, migrationID)
	if err != nil {
		return nil, err
	}
	for i, d := range dss {
		i := i
		stages = append(stages, stage.Stage[any]{
			ProgressMsg: fmt.Sprintf("Migrating %s: %s", d.Type, d.Name),
			SuccessMsg:  fmt.Sprintf("Migrated %s: %s ...", d.Type, d.Name),
			FailureMsg:  fmt.Sprintf("Failed migrating %s: %s ...", d.Type, d.Name),
			Func: func(ct context.Context, status stage.Statuser[any]) (any, error) {
				var execErr error
			StatusReaderLoop:
				for {
					if ctx.Err() != nil {
						if errors.Is(err, context.DeadlineExceeded) {
							execErr = timeoutErr
							break StatusReaderLoop
						}
						execErr = fmt.Errorf("migration failed: %w", err)
						break StatusReaderLoop
					}
					generalStatus, err := fetchMigrationStatus(ctx, ci, migrationID)
					if err != nil {
						execErr = fmt.Errorf("reading migration status: %w", err)
						break StatusReaderLoop
					}
					switch Status(generalStatus) {
					case StatusComplete:
						return nil, nil
					case StatusFailed:
						errs, err := fetchMigrationErrors(ctx, ci, migrationID)
						if err != nil {
							execErr = fmt.Errorf("fetching migration errors: %w", err)
							break StatusReaderLoop
						}
						execErr = errors.New(errs)
						break StatusReaderLoop
					case StatusCanceled, StatusCanceling:
						execErr = errors2.ErrUserCancelled
						break StatusReaderLoop
					}
					q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.migrations[%d]') FROM %s WHERE __key= '%s'`, i, StatusMapName, migrationID)
					res, err := ci.Client().SQL().Execute(ctx, q)
					if err != nil {
						execErr = err
						break StatusReaderLoop
					}
					iter, err := res.Iterator()
					if err != nil {
						execErr = err
						break StatusReaderLoop
					}
					if iter.HasNext() {
						row, err := iter.Next()
						if err != nil {
							execErr = err
							break StatusReaderLoop
						}
						rowStr, err := row.Get(0)
						if err != nil {
							execErr = err
							break StatusReaderLoop
						}
						var m DSMigrationStatus
						if err = json.Unmarshal(rowStr.(serialization.JSON), &m); err != nil {
							execErr = err
							break StatusReaderLoop
						}
						status.SetProgress(m.CompletionPercentage)
						switch m.Status {
						case StatusComplete:
							return nil, nil
						case StatusFailed:
							return nil, stage.IgnoreError(errors.New(m.Error))
						case StatusCanceled:
							execErr = errors2.ErrUserCancelled
							break StatusReaderLoop
						}
					}
					time.Sleep(1 * time.Second)
				}
				if execErr != nil {
					status.SetText(execErr.Error())
				}
				return nil, execErr
			},
		})
	}
	return stages, nil
}

func dataStructuresToBeMigrated(ctx context.Context, ec plug.ExecContext, migrationID string) ([]DataStructureInfo, error) {
	var dss []DataStructureInfo
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf(`SELECT this FROM %s WHERE __key= '%s'`, StatusMapName, migrationID)
	res, err := ci.Client().SQL().Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	it, err := res.Iterator()
	if err != nil {
		return nil, err
	}
	if it.HasNext() { // single iteration is enough that we are reading single result for a single migration
		row, err := it.Next()
		if err != nil {
			return nil, err
		}
		r, err := row.Get(0)
		if err != nil {
			return nil, err
		}
		var status OverallMigrationStatus
		if err = json.Unmarshal(r.(serialization.JSON), &status); err != nil {
			return nil, err
		}
		for _, m := range status.Migrations {
			dss = append(dss, DataStructureInfo{
				Name: m.Name,
				Type: m.Type,
			})
		}
	}
	return dss, nil
}

func saveMemberLogs(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, migrationID string) error {
	for _, m := range ci.OrderedMembers() {
		l, err := ci.Client().GetList(ctx, DebugLogsListPrefix+m.UUID.String())
		if err != nil {
			return err
		}
		logs, err := l.GetAll(ctx)
		if err != nil {
			return err
		}
		for _, line := range logs {
			ec.Logger().Info(fmt.Sprintf("[%s_%s] %s", migrationID, m.UUID.String(), line.(string)))
		}
	}
	return nil
}

func saveReportToFile(ctx context.Context, ci *hazelcast.ClientInternal, migrationID, fileName string) error {
	report, err := fetchMigrationReport(ctx, ci, migrationID)
	if err != nil {
		return err
	}
	f, err := os.Create(fmt.Sprintf(fileName))
	if err != nil {
		return err
	}
	defer f.Close()
	return os.WriteFile(fileName, []byte(report), 0600)
}

func waitForMigrationToBeCreated(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) error {
	for {
		statusMap, err := ci.Client().GetMap(ctx, StatusMapName)
		if err != nil {
			return err
		}
		ok, err := statusMap.ContainsKey(ctx, migrationID)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
}

type OverallMigrationStatus struct {
	Status               Status              `json:"status"`
	Logs                 []string            `json:"logs"`
	Errors               []string            `json:"errors"`
	Report               string              `json:"report"`
	CompletionPercentage float32             `json:"completionPercentage"`
	Migrations           []DSMigrationStatus `json:"migrations"`
}

type DataStructureInfo struct {
	Name string
	Type string
}

type DSMigrationStatus struct {
	Name                 string  `json:"name"`
	Type                 string  `json:"type"`
	Status               Status  `json:"status"`
	CompletionPercentage float32 `json:"completionPercentage"`
	Error                string  `json:"error"`
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

func fetchMigrationReport(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.report') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
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

func fetchMigrationErrors(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.errors') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	res, err := ci.Client().SQL().Execute(ctx, q)
	if err != nil {
		return "", err
	}
	it, err := res.Iterator()
	if err != nil {
		return "", err
	}
	var errs []string
	for it.HasNext() { // single iteration is enough that we are reading single result for a single migration
		row, err := it.Next()
		if err != nil {
			return "", err
		}
		r, err := row.Get(0)
		var m string
		if err = json.Unmarshal(r.(serialization.JSON), &m); err != nil {
			return "", err
		}
		errs = append(errs, m)
	}
	return strings.Join(errs, "\n"), nil
}

func finalizeMigration(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, migrationID, reportOutputDir string) error {
	err := saveMemberLogs(ctx, ec, ci, migrationID)
	if err != nil {
		return err
	}
	var name string
	if reportOutputDir == "" {
		name = fmt.Sprintf("migration_report_%s.txt", migrationID)
	}
	err = saveReportToFile(ctx, ci, migrationID, name)
	if err != nil {
		return fmt.Errorf("saving report to file: %w", err)
	}
	return nil
}
