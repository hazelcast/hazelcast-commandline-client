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
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "status", f: statusTest},
		{name: "noMigrationsStatus", f: noMigrationsStatusTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func noMigrationsStatusTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		var wg sync.WaitGroup
		wg.Add(1)
		go tcx.WithReset(func() {
			defer wg.Done()
			tcx.CLC().Execute(ctx, "status")
		})
		wg.Wait()
		tcx.AssertStdoutContains("there are no migrations are in progress on migration cluster")
	})
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
		statusRunner(t, mID, tcx, ctx)
		wg.Wait()
		tcx.AssertStdoutContains(`
Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.

 OK   [1/2] Connected to the migration cluster.
first message
last message
Completion Percentage: 12.123000
status report
imap5	IMap	FAILED	2023-01-01 00:00:00	141	1000	14.1
 OK   [2/2] Fetched migration status.

OK`)
	})
}

func preStatusRunner(t *testing.T, tcx it.TestContext, ctx context.Context) string {
	mID := migration.MakeMigrationID()
	l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
	m := MustValue(json.Marshal(migration.MigrationInProgress{
		MigrationID: mID,
	}))
	ok := MustValue(l.Add(ctx, serialization.JSON(m)))
	require.Equal(t, true, ok)
	return mID
}

func statusRunner(t *testing.T, migrationID string, tcx it.TestContext, ctx context.Context) {
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.MakeStatusMapName(migrationID)))
	topic := MustValue(tcx.Client.GetTopic(ctx, migration.MakeUpdateTopicName(migrationID)))
	setState(ctx, topic, statusMap, migration.StatusInProgress, "first message")
	setState(ctx, topic, statusMap, migration.StatusFailed, "last message")
	l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
	m := MustValue(json.Marshal(migration.MigrationInProgress{
		MigrationID: migrationID,
	}))
	ok := MustValue(l.Remove(ctx, serialization.JSON(m)))
	require.Equal(t, true, ok)
}

func setState(ctx context.Context, updateTopic *hazelcast.Topic, statusMap *hazelcast.Map, status migration.Status, msg string) {
	startTime := MustValue(time.Parse(time.RFC3339, "2023-01-01T00:00:00Z"))
	st := MustValue(json.Marshal(migration.MigrationStatus{
		Status:               status,
		Report:               "status report",
		CompletionPercentage: 12.123,
		Migrations: []migration.Migration{
			{
				Name:                 "imap5",
				Type:                 "IMap",
				Status:               status,
				StartTimestamp:       startTime,
				EntriesMigrated:      141,
				TotalEntries:         1000,
				CompletionPercentage: 14.1,
			},
		},
	}))
	message := MustValue(json.Marshal(migration.UpdateMessage{Status: status, Message: msg, CompletionPercentage: 80}))
	Must(statusMap.Set(ctx, migration.StatusMapEntryName, serialization.JSON(st)))
	Must(updateTopic.Publish(ctx, serialization.JSON(message)))
}
