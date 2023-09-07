//go:build migration

package migration_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands/migration"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "status", f: statusTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func statusTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		mID := preStatusRunner(t, tcx, ctx)
		var wg sync.WaitGroup
		wg.Add(1)
		go tcx.WithReset(func() {
			defer wg.Done()
			Must(tcx.CLC().Execute(ctx, "status"))
		})
		time.Sleep(1 * time.Second) // give time to status command to register its topic listener
		statusRunner(mID, tcx, ctx)
		wg.Wait()
		tcx.AssertStdoutContains(`
Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.

 OK   [1/2] Connected to the migration cluster.
first message
last message
status report
imap5	IMap	FAILED	2023-01-01 00:00:00	141	1000	14.1
 OK   [2/2] Fetched migration status.

OK`)
	})
}

func preStatusRunner(t *testing.T, tcx it.TestContext, ctx context.Context) string {
	mID := migration.MakeMigrationID()
	l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
	ok := MustValue(l.Add(ctx, migration.MigrationInProgress{
		MigrationID: mID,
	}))
	require.Equal(t, true, ok)
	return mID
}

func statusRunner(migrationID string, tcx it.TestContext, ctx context.Context) {
	startTime := MustValue(time.Parse(time.RFC3339, "2023-01-01T00:00:00Z"))
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.MakeStatusMapName(migrationID)))
	b := MustValue(json.Marshal(migration.MigrationStatus{
		Status: migration.StatusInProgress,
		Report: "status report",
		Migrations: []migration.Migration{
			{
				Name:                 "imap5",
				Type:                 "IMap",
				Status:               migration.StatusInProgress,
				StartTimestamp:       startTime,
				EntriesMigrated:      121,
				TotalEntries:         1000,
				CompletionPercentage: 12.1,
			},
		},
	}))
	Must(statusMap.Set(ctx, migration.StatusMapEntryName, serialization.JSON(b)))
	topic := MustValue(tcx.Client.GetTopic(ctx, migration.MakeUpdateTopicName(migrationID)))
	Must(topic.Publish(ctx, migration.UpdateMessage{Status: migration.StatusInProgress, Message: "first message"}))
	b = MustValue(json.Marshal(migration.MigrationStatus{
		Status: migration.StatusFailed,
		Report: "status report",
		Migrations: []migration.Migration{
			{
				Name:                 "imap5",
				Type:                 "IMap",
				Status:               migration.StatusFailed,
				StartTimestamp:       startTime,
				EntriesMigrated:      141,
				TotalEntries:         1000,
				CompletionPercentage: 14.1,
			},
		},
	}))
	Must(statusMap.Set(ctx, migration.StatusMapEntryName, serialization.JSON(b))) // update status map
	Must(topic.Publish(ctx, migration.UpdateMessage{Status: migration.StatusFailed, Message: "last message"}))
}
