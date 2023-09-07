package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type MigrationStatus struct {
	Status               Status      `json:"status"`
	Logs                 []string    `json:"logs"`
	Errors               []string    `json:"errors"`
	Report               string      `json:"report"`
	Migrations           []Migration `json:"migrations"`
	CompletionPercentage float32     `json:"completionPercentage"`
}

type Migration struct {
	Name                 string    `json:"name"`
	Type                 string    `json:"type"`
	Status               Status    `json:"status"`
	StartTimestamp       time.Time `json:"startTimestamp"`
	EntriesMigrated      int       `json:"entriesMigrated"`
	TotalEntries         int       `json:"totalEntries"`
	CompletionPercentage float64   `json:"completionPercentage"`
}

type UpdateMessage struct {
	Status               Status  `json:"status"`
	CompletionPercentage float32 `json:"completionPercentage"`
	Message              string  `json:"message"`
}

func readMigrationStatus(ctx context.Context, statusMap *hazelcast.Map) (*MigrationStatus, error) {
	v, err := statusMap.Get(ctx, StatusMapEntryName)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	var b []byte
	if vv, ok := v.(string); ok {
		b = []byte(vv)
	} else if vv, ok := v.(serialization.JSON); ok {
		b = vv
	} else {
		return nil, fmt.Errorf("invalid status value")
	}
	var ms MigrationStatus
	if err := json.Unmarshal(b, &ms); err != nil {
		return nil, fmt.Errorf("unmarshaling status: %w", err)
	}
	return &ms, nil
}
