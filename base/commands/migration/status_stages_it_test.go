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
Completion Percentage: 60.000000
last message
Completion Percentage: 100.000000
status report
 OK   [2/2] Fetched migration status.

OK`)
	})
}

func preStatusRunner(t *testing.T, tcx it.TestContext, ctx context.Context) string {
	// create a migration in the __datamigrations_in_progress list
	mID := migration.MakeMigrationID()
	l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
	m := MustValue(json.Marshal(migration.MigrationInProgress{
		MigrationID: mID,
	}))
	ok := MustValue(l.Add(ctx, serialization.JSON(m)))
	require.Equal(t, true, ok)
	// create a record in the status map
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.MakeStatusMapName(mID)))
	st := MustValue(json.Marshal(migration.MigrationStatus{
		Status:               migration.StatusInProgress,
		Report:               "status report",
		CompletionPercentage: 60,
	}))
	Must(statusMap.Set(ctx, migration.StatusMapEntryName, serialization.JSON(st)))
	return mID
}

func statusRunner(t *testing.T, migrationID string, tcx it.TestContext, ctx context.Context) {
	// publish the first message in the update topic
	updateTopic := MustValue(tcx.Client.GetTopic(ctx, migration.MakeUpdateTopicName(migrationID)))
	msg := MustValue(json.Marshal(migration.UpdateMessage{Status: migration.StatusInProgress, Message: "first message", CompletionPercentage: 60}))
	Must(updateTopic.Publish(ctx, serialization.JSON(msg)))
	// create a terminal record in status map
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.MakeStatusMapName(migrationID)))
	st := MustValue(json.Marshal(migration.MigrationStatus{
		Status:               migration.StatusComplete,
		Report:               "status report",
		CompletionPercentage: 100,
	}))
	Must(statusMap.Set(ctx, migration.StatusMapEntryName, serialization.JSON(st)))
	// publish the second message in the update topic
	msg = MustValue(json.Marshal(migration.UpdateMessage{Status: migration.StatusComplete, Message: "last message", CompletionPercentage: 100}))
	Must(updateTopic.Publish(ctx, serialization.JSON(msg)))
	// remove the migration from the __datamigrations_in_progress list
	l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
	m := MustValue(json.Marshal(migration.MigrationInProgress{
		MigrationID: migrationID,
	}))
	ok := MustValue(l.Remove(ctx, serialization.JSON(m)))
	require.Equal(t, true, ok)
}
