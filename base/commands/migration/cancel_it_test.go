//go:build std || migration

package migration_test

import (
	"context"
	"encoding/json"
	"os"
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
		err := tcx.CLC().Execute(ctx, "cancel")
		require.Contains(t, err.Error(), "there are no migrations in progress")
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
		tcx.WithReset(func() {
			Must(tcx.CLC().Execute(ctx, "cancel"))
		})
		MustValue(l.Remove(ctx, serialization.JSON(m)))
		tcx.AssertStdoutContains(`Migration canceled successfully.`)
	})
}
