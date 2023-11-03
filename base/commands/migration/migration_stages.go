//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

var timeoutErr = fmt.Errorf("migration could not be completed: reached timeout while reading status: "+
	"please ensure that you are using Hazelcast's migration cluster distribution and your DMT configuration points to that cluster: %w",
	context.DeadlineExceeded)

var migrationStatusNotFoundErr = fmt.Errorf("migration status not found")

func createMigrationStages(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, migrationID string) ([]stage.Stage[any], error) {
	childCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := WaitForMigrationToBeInProgress(childCtx, ci, migrationID); err != nil {
		return nil, fmt.Errorf("waiting migration to be created: %w", err)
	}
	var stages []stage.Stage[any]
	dss, err := getDataStructuresToBeMigrated(ctx, ec, migrationID)
	if err != nil {
		return nil, err
	}
	for i, d := range dss {
		i := i
		stages = append(stages, stage.Stage[any]{
			ProgressMsg: fmt.Sprintf("Migrating %s: %s", d.Type, d.Name),
			SuccessMsg:  fmt.Sprintf("Migrated %s: %s", d.Type, d.Name),
			FailureMsg:  fmt.Sprintf("Failed migrating %s: %s", d.Type, d.Name),
			Func: func(ct context.Context, status stage.Statuser[any]) (any, error) {
				var execErr error
			statusReaderLoop:
				for {
					if ctx.Err() != nil {
						if errors.Is(err, context.DeadlineExceeded) {
							execErr = timeoutErr
							break statusReaderLoop
						}
						execErr = fmt.Errorf("migration failed: %w", err)
						break statusReaderLoop
					}
					generalStatus, err := fetchMigrationStatus(ctx, ci, migrationID)
					if err != nil {
						execErr = fmt.Errorf("reading migration status: %w", err)
						break statusReaderLoop
					}
					switch Status(generalStatus) {
					case StatusStarted:
						break statusReaderLoop
					case StatusComplete:
						return nil, nil
					case StatusFailed:
						errs, err := fetchMigrationErrors(ctx, ci, migrationID)
						if err != nil {
							execErr = fmt.Errorf("fetching migration errors: %w", err)
							break statusReaderLoop
						}
						execErr = errors.New(errs)
						break statusReaderLoop
					case StatusCanceled, StatusCanceling:
						execErr = clcerrors.ErrUserCancelled
						break statusReaderLoop
					}
					q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.migrations[%d]') FROM %s WHERE __key= '%s'`, i, StatusMapName, migrationID)
					res, err := ci.Client().SQL().Execute(ctx, q)
					if err != nil {
						execErr = err
						break statusReaderLoop
					}
					iter, err := res.Iterator()
					if err != nil {
						execErr = err
						break statusReaderLoop
					}
					if iter.HasNext() {
						row, err := iter.Next()
						if err != nil {
							execErr = err
							break statusReaderLoop
						}
						rowStr, err := row.Get(0)
						if err != nil {
							execErr = err
							break statusReaderLoop
						}
						var m DataStructureMigrationStatus
						if err = json.Unmarshal(rowStr.(serialization.JSON), &m); err != nil {
							execErr = err
							break statusReaderLoop
						}
						status.SetProgress(m.CompletionPercentage)
						switch m.Status {
						case StatusComplete:
							return nil, nil
						case StatusFailed:
							return nil, stage.IgnoreError(errors.New(m.Error))
						case StatusCanceled:
							execErr = clcerrors.ErrUserCancelled
							break statusReaderLoop
						}
					}
					time.Sleep(1 * time.Second)
				}
				return nil, execErr
			},
		})
	}
	return stages, nil
}

func getDataStructuresToBeMigrated(ctx context.Context, ec plug.ExecContext, migrationID string) ([]DataStructureInfo, error) {
	var dss []DataStructureInfo
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf(`SELECT this FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	r, err := querySingleRow(ctx, ci, q)
	if err != nil {
		return nil, err
	}
	var status OverallMigrationStatus
	if err = json.Unmarshal(r.(serialization.JSON), &status); err != nil {
		return nil, err
	}
	if len(status.Migrations) == 0 {
		return nil, errors.New("no datastructures found to migrate")
	}
	for _, m := range status.Migrations {
		dss = append(dss, DataStructureInfo{
			Name: m.Name,
			Type: m.Type,
		})
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
	if report == "" {
		return nil
	}
	return os.WriteFile(fileName, []byte(report), 0600)
}

func WaitForMigrationToBeInProgress(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) error {
	for {
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
				return fmt.Errorf("migration failed and dmt cannot fetch migration errors: %w", err)
			}
			return errors.New(errs)
		}
		if Status(status) == StatusInProgress || Status(status) == StatusComplete {
			return nil
		}
	}
}

type OverallMigrationStatus struct {
	Status               Status                         `json:"status"`
	Logs                 []string                       `json:"logs"`
	Errors               []string                       `json:"errors"`
	Report               string                         `json:"report"`
	CompletionPercentage float32                        `json:"completionPercentage"`
	Migrations           []DataStructureMigrationStatus `json:"migrations"`
}

type DataStructureInfo struct {
	Name string
	Type string
}

type DataStructureMigrationStatus struct {
	Name                 string  `json:"name"`
	Type                 string  `json:"type"`
	Status               Status  `json:"status"`
	CompletionPercentage float32 `json:"completionPercentage"`
	Error                string  `json:"error"`
}

func fetchMigrationStatus(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.status') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	r, err := querySingleRow(ctx, ci, q)
	if err != nil {
		return "", migrationStatusNotFoundErr
	}
	return strings.TrimSuffix(strings.TrimPrefix(string(r.(serialization.JSON)), `"`), `"`), nil
}

func fetchMigrationReport(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.report') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	r, err := querySingleRow(ctx, ci, q)
	if err != nil {
		return "", fmt.Errorf("migration report cannot be found: %w", err)
	}
	return strings.ReplaceAll(string(r.(serialization.JSON)), `\"`, ``), nil
}

func fetchMigrationErrors(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.errors' WITH WRAPPER) FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	row, err := querySingleRow(ctx, ci, q)
	if err != nil {
		return "", err
	}
	var errs []string
	err = json.Unmarshal(row.(serialization.JSON), &errs)
	if err != nil {
		return "", err
	}
	return "* " + strings.Join(errs, "\n* "), nil
}

func finalizeMigration(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, migrationID, reportOutputDir string) error {
	err := saveMemberLogs(ctx, ec, ci, migrationID)
	if err != nil {
		return err
	}
	outFile := filepath.Join(reportOutputDir, fmt.Sprintf("migration_report_%s.txt", migrationID))
	err = saveReportToFile(ctx, ci, migrationID, outFile)
	if err != nil {
		return fmt.Errorf("saving report to file: %w", err)
	}
	return nil
}

func querySingleRow(ctx context.Context, ci *hazelcast.ClientInternal, query string) (any, error) {
	res, err := ci.Client().SQL().Execute(ctx, query)
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
		r, err := row.Get(0)
		if err != nil {
			return "", err
		}
		return r, nil
	}
	return nil, errors.New("no rows found")
}
