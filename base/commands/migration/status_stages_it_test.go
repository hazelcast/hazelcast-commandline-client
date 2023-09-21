//go:build std || migration

package migration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
		time.Sleep(1 * time.Second)
		statusRunner(t, mID, tcx, ctx)
		wg.Wait()
		tcx.AssertStdoutContains("OK  Connected to the migration cluster.")
		tcx.WithReset(func() {
			f := fmt.Sprintf("migration_report_%s.txt", mID)
			require.Equal(t, true, fileExists(f))
			Must(os.Remove(f))
		})
	})
}

func preStatusRunner(t *testing.T, tcx it.TestContext, ctx context.Context) string {
	createMapping(ctx, tcx)
	// create a migration in the __datamigrations_in_progress list
	mID := migration.MakeMigrationID()
	l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
	m := MustValue(json.Marshal(migration.MigrationInProgress{
		MigrationID: mID,
	}))
	ok := MustValue(l.Add(ctx, serialization.JSON(m)))
	require.Equal(t, true, ok)
	// create a record in the status map
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := MustValue(os.ReadFile("testdata/start/migration_success_initial.json"))
	Must(statusMap.Set(ctx, mID, serialization.JSON(b)))
	return mID
}

func statusRunner(t *testing.T, migrationID string, tcx it.TestContext, ctx context.Context) {
	// create a terminal record in status map
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := MustValue(os.ReadFile("testdata/start/migration_success_completed.json"))
	Must(statusMap.Set(ctx, migrationID, serialization.JSON(b)))
	// remove the migration from the __datamigrations_in_progress list
	l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
	m := MustValue(json.Marshal(migration.MigrationInProgress{
		MigrationID: migrationID,
	}))
	require.Equal(t, true, MustValue(l.Remove(ctx, serialization.JSON(m))))
}
