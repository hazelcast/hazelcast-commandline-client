//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type MigrationStatusTotal struct {
	Status               Status               `json:"status"`
	Logs                 []string             `json:"logs"`
	Errors               []string             `json:"errors"`
	Report               string               `json:"report"`
	CompletionPercentage float32              `json:"completionPercentage"`
	Migrations           []MigrationStatusRow `json:"migrations"`
}

type DataStructureInfo struct {
	Name string
	Type string
}

type MigrationStatusRow struct {
	Name                 string  `json:"name"`
	Type                 string  `json:"type"`
	Status               Status  `json:"status"`
	CompletionPercentage float32 `json:"completionPercentage"`
	Error                string  `json:"error"`
}

func readMigrationStatus(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.status') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	res, err := ci.Client().SQL().Execute(ctx, q)
	if err != nil {
		return "", err
	}
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

func readMigrationReport(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.report') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	res, err := ci.Client().SQL().Execute(ctx, q)
	if err != nil {
		return "", err
	}
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

func readMigrationErrors(ctx context.Context, ci *hazelcast.ClientInternal, migrationID string) (string, error) {
	q := fmt.Sprintf(`SELECT JSON_QUERY(this, '$.errors') FROM %s WHERE __key='%s'`, StatusMapName, migrationID)
	res, err := ci.Client().SQL().Execute(ctx, q)
	if err != nil {
		return "", err
	}
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
