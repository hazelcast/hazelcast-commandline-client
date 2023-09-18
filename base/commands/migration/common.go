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

type MigrationStatus struct {
	Status               Status   `json:"status"`
	Logs                 []string `json:"logs"`
	Errors               []string `json:"errors"`
	Report               string   `json:"report"`
	CompletionPercentage float32  `json:"completionPercentage"`
}

type UpdateMessage struct {
	Status               Status  `json:"status"`
	CompletionPercentage float32 `json:"completionPercentage"`
	Message              string  `json:"message"`
}

var ErrInvalidStatus = errors.New("invalid status value")

func readMigrationStatus(ctx context.Context, statusMap *hazelcast.Map) (*MigrationStatus, error) {
	v, err := statusMap.Get(ctx, StatusMapEntryName)
	if err != nil {
		return nil, fmt.Errorf("getting status: %w", err)
	}
	if v == nil {
		return nil, ErrInvalidStatus
	}
	var b []byte
	if vv, ok := v.(string); ok {
		b = []byte(vv)
	} else if vv, ok := v.(serialization.JSON); ok {
		b = vv
	} else {
		return nil, ErrInvalidStatus
	}
	var ms MigrationStatus
	if err := json.Unmarshal(b, &ms); err != nil {
		return nil, fmt.Errorf("parsing migration status: %w", err)
	}
	return &ms, nil
}
