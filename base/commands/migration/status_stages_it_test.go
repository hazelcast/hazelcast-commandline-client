//go:build std || migration

package migration_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands/migration"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	hz "github.com/hazelcast/hazelcast-go-client"
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
		createMapping(ctx, tcx)
		var wg sync.WaitGroup
		wg.Add(1)
		var execErr error
		go tcx.WithReset(func() {
			defer wg.Done()
			execErr = tcx.CLC().Execute(ctx, "status")
		})
		wg.Wait()
		require.Contains(t, execErr.Error(), "finding migration in progress: no rows found")
	})
}

func statusTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		ci := hz.NewClientInternal(tcx.Client)
		createMapping(ctx, tcx)
		createMemberLogs(t, ctx, ci)
		defer removeMembersLogs(ctx, ci)
		outDir := MustValue(os.MkdirTemp("", "clc-"))
		mID := setStatusInProgress(tcx, ctx)
		var wg sync.WaitGroup
		wg.Add(1)
		go tcx.WithReset(func() {
			defer wg.Done()
			Must(tcx.CLC().Execute(ctx, "status", "-o", outDir))
		})
		time.Sleep(1 * time.Second)
		setStatusCompleted(mID, tcx, ctx)
		wg.Wait()
		tcx.AssertStdoutContains("Connected to the migration cluster")
		tcx.WithReset(func() {
			f := paths.Join(outDir, fmt.Sprintf("migration_report_%s.txt", mID))
			require.Equal(t, true, paths.Exists(f))
			Must(os.Remove(f))
			b := MustValue(os.ReadFile(paths.ResolveLogPath("test")))
			for _, m := range ci.OrderedMembers() {
				require.Contains(t, string(b), fmt.Sprintf("[%s_%s] log1", mID, m.UUID.String()))
				require.Contains(t, string(b), fmt.Sprintf("[%s_%s] log2", mID, m.UUID.String()))
				require.Contains(t, string(b), fmt.Sprintf("[%s_%s] log3", mID, m.UUID.String()))
			}
		})
	})
}

func setStatusInProgress(tcx it.TestContext, ctx context.Context) string {
	mID := migrationIDFunc()
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := MustValue(os.ReadFile("testdata/start/migration_success_initial.json"))
	Must(statusMap.Set(ctx, mID, serialization.JSON(b)))
	return mID
}

func setStatusCompleted(migrationID string, tcx it.TestContext, ctx context.Context) {
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := MustValue(os.ReadFile("testdata/start/migration_success_completed.json"))
	Must(statusMap.Set(ctx, migrationID, serialization.JSON(b)))
}

func createMemberLogs(t *testing.T, ctx context.Context, ci *hz.ClientInternal) {
	for _, m := range ci.OrderedMembers() {
		l := MustValue(ci.Client().GetList(ctx, migration.DebugLogsListPrefix+m.UUID.String()))
		require.Equal(t, true, MustValue(l.Add(ctx, "log1")))
		require.Equal(t, true, MustValue(l.Add(ctx, "log2")))
		require.Equal(t, true, MustValue(l.Add(ctx, "log3")))
	}
}

func removeMembersLogs(ctx context.Context, ci *hz.ClientInternal) {
	for _, m := range ci.OrderedMembers() {
		l := MustValue(ci.Client().GetList(ctx, migration.DebugLogsListPrefix+m.UUID.String()))
		Must(l.Destroy(ctx))
	}
}
