//go:build std || migration

package migration_test

import (
	"context"
	"os"
	"sync"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands/migration"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	hz "github.com/hazelcast/hazelcast-go-client"
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
		createMapping(ctx, tcx)
		err := tcx.CLC().Execute(ctx, "cancel")
		require.Contains(t, err.Error(), "finding migration in progress: no rows found")
	})
}

func cancelTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		ci := hz.NewClientInternal(tcx.Client)
		mID := migrationIDFunc()
		createMapping(ctx, tcx)
		setStatusInProgress(tcx, ctx)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			Must(tcx.CLC().Execute(ctx, "cancel"))
		}()
		cq := MustValue(ci.Client().GetQueue(ctx, migration.CancelQueue))
		MustValue(cq.Poll(ctx))
		setStatusCancelling(mID, tcx, ctx)
		wg.Wait()
		statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
		MustValue(statusMap.Remove(ctx, mID))
		tcx.AssertStdoutContains(`Migration canceled successfully.`)
	})
}

func setStatusCancelling(migrationID string, tcx it.TestContext, ctx context.Context) {
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := MustValue(os.ReadFile("testdata/cancel/migration_cancelling.json"))
	Must(statusMap.Set(ctx, migrationID, serialization.JSON(b)))
}
