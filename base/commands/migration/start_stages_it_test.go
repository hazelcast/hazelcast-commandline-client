//go:build std || migration

package migration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/stretchr/testify/require"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands/migration"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestMigrationStages(t *testing.T) {
	testCases := []struct {
		name                string
		statusMapStateFiles []string
		expectedErr         error
	}{
		{
			name: "successful",
			statusMapStateFiles: []string{
				"testdata/start/migration_success_initial.json",
				"testdata/start/migration_success_completed.json",
			},
		},
		{
			name: "failure",
			statusMapStateFiles: []string{
				"testdata/start/migration_success_initial.json",
				"testdata/start/migration_success_failure.json",
			},
			expectedErr: errors.New("Failed migrating IMAP: imap5: * some error\n* another error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			startMigrationTest(t, tc.expectedErr, tc.statusMapStateFiles)
		})
	}
}

func startMigrationTest(t *testing.T, expectedErr error, statusMapStateFiles []string) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		ci := hz.NewClientInternal(tcx.Client)
		createMapping(ctx, tcx)
		createMemberLogs(t, ctx, ci)
		defer removeMembersLogs(ctx, ci)
		outDir := MustValue(os.MkdirTemp("", "clc-"))
		var wg sync.WaitGroup
		wg.Add(1)
		var execErr error
		go tcx.WithReset(func() {
			defer wg.Done()
			execErr = tcx.CLC().Execute(ctx, "start", "testdata/dmt_config", "--yes", "-o", outDir)
		})
		c := make(chan string)
		go findMigrationID(ctx, tcx, c)
		mID := <-c
		wg.Add(1)
		go migrationRunner(t, ctx, tcx, mID, &wg, statusMapStateFiles)
		wg.Wait()
		if expectedErr == nil {
			require.Equal(t, nil, execErr)
		} else {
			require.Contains(t, execErr.Error(), expectedErr.Error())
		}
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

func migrationRunner(t *testing.T, ctx context.Context, tcx it.TestContext, migrationID string, wg *sync.WaitGroup, statusMapStateFiles []string) {
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	for _, f := range statusMapStateFiles {
		b := MustValue(os.ReadFile(f))
		it.Eventually(t, func() bool {
			return statusMap.Set(ctx, migrationID, serialization.JSON(b)) == nil
		})
	}
	wg.Done()
}

func createMapping(ctx context.Context, tcx it.TestContext) {
	mSQL := fmt.Sprintf(`CREATE MAPPING IF NOT EXISTS %s TYPE IMap OPTIONS('keyFormat'='varchar', 'valueFormat'='json')`, migration.StatusMapName)
	MustValue(tcx.Client.SQL().Execute(ctx, mSQL))
}

func findMigrationID(ctx context.Context, tcx it.TestContext, c chan string) {
	q := MustValue(tcx.Client.GetQueue(ctx, migration.StartQueueName))
	var b migration.ConfigBundle
	for {
		v := MustValue(q.PollWithTimeout(ctx, time.Second))
		if v != nil {
			Must(json.Unmarshal(v.(serialization.JSON), &b))
			c <- b.MigrationID
			break
		}
	}
}
