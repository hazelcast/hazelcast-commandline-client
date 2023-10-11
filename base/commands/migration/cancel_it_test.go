//go:build migration

package migration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands/migration"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/stretchr/testify/require"
)

func TestCancel(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "cancel", f: cancelTest},
		{name: "noMigrations", f: noMigrationsCancelTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func noMigrationsCancelTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		var wg sync.WaitGroup
		wg.Add(1)
		go tcx.WithReset(func() {
			defer wg.Done()
			tcx.CLC().Execute(ctx, "cancel")
		})
		wg.Wait()
		tcx.AssertStdoutContains("there are no migrations are in progress on migration cluster")
	})
}

func cancelTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		mID := migration.MakeMigrationID()
		createMapping(ctx, tcx)
		statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
		b := MustValue(os.ReadFile("testdata/cancel/migration_cancelling.json"))
		Must(statusMap.Set(ctx, mID, serialization.JSON(b)))
		l := MustValue(tcx.Client.GetList(ctx, migration.MigrationsInProgressList))
		m := MustValue(json.Marshal(migration.MigrationInProgress{
			MigrationID: mID,
		}))
		ok := MustValue(l.Add(ctx, serialization.JSON(m)))
		require.Equal(t, true, ok)
		var wg sync.WaitGroup
		wg.Add(1)
		go tcx.WithReset(func() {
			defer wg.Done()
			Must(tcx.CLC().Execute(ctx, "cancel"))
		})
		wg.Wait()
		MustValue(l.Remove(ctx, serialization.JSON(m)))
		tcx.AssertStdoutContains(`Migration canceled successfully.`)
	})
}

func createMapping(ctx context.Context, tcx it.TestContext) {
	mSQL := fmt.Sprintf(`CREATE MAPPING IF NOT EXISTS %s TYPE IMap OPTIONS('keyFormat'='varchar', 'valueFormat'='json')`, migration.StatusMapName)
	MustValue(tcx.Client.SQL().Execute(ctx, mSQL))
}
