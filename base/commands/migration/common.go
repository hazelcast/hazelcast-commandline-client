//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

var timeoutErr = errors.New("migration could not be completed: reached timeout while reading status: " +
	"please ensure that you are using Hazelcast's migration cluster distribution and your DMT configuration points to that cluster")

type MigrationInProgress struct {
	MigrationID string `json:"id"`
}

func findMigrationInProgress(ctx context.Context, ci *hazelcast.ClientInternal) (MigrationInProgress, error) {
	var mip MigrationInProgress
	q := fmt.Sprintf("SELECT this FROM %s WHERE JSON_VALUE(this, '$.status') IN('STARTED', 'IN_PROGRESS', 'CANCELING')", StatusMapName)
	r, err := querySingleRow(ctx, ci, q)
	if err != nil {
		return mip, fmt.Errorf("finding migration in progress: %w", err)
	}
	rr, err := r.Get(0)
	if err != nil {
		return mip, fmt.Errorf("finding migration in progress: %w", err)
	}
	m := rr.(serialization.JSON)
	if err = json.Unmarshal(m, &mip); err != nil {
		return mip, fmt.Errorf("parsing migration in progress: %w", err)
	}
	return mip, nil
}
