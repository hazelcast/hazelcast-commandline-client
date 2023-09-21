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

func TestMigrationStages(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "start_Successful", f: startTest_Successful},
		{name: "start_Failure", f: startTest_Failure},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func startTest_Successful(t *testing.T) {
	startTest(t, successfulRunner, "OK Migration completed successfully.")
}

func startTest_Failure(t *testing.T) {
	startTest(t, failureRunner, "ERROR Failed migrating IMAP: imap5 ...: some error")
}

func startTest(t *testing.T, runnerFunc func(context.Context, it.TestContext, string, *sync.WaitGroup), expectedOutput string) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		var wg sync.WaitGroup
		wg.Add(1)
		go tcx.WithReset(func() {
			defer wg.Done()
			tcx.CLC().Execute(ctx, "start", "dmt-config", "--yes")
		})
		c := make(chan string, 1)
		wg.Add(1)
		go findMigrationID(ctx, tcx, c)
		mID := <-c
		wg.Done()
		wg.Add(1)
		go runnerFunc(ctx, tcx, mID, &wg)
		wg.Wait()
		tcx.AssertStdoutContains(expectedOutput)
		tcx.WithReset(func() {
			f := fmt.Sprintf("migration_report_%s.txt", mID)
			require.Equal(t, true, fileExists(f))
			Must(os.Remove(f))
		})
	})
}

func successfulRunner(ctx context.Context, tcx it.TestContext, migrationID string, wg *sync.WaitGroup) {
	mSQL := fmt.Sprintf(`CREATE MAPPING IF NOT EXISTS %s TYPE IMap OPTIONS('keyFormat'='varchar', 'valueFormat'='json')`, migration.StatusMapName)
	MustValue(tcx.Client.SQL().Execute(ctx, mSQL))
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := MustValue(os.ReadFile("testdata/start/migration_success_initial.json"))
	MustValue(statusMap.Put(ctx, migrationID, serialization.JSON(b)))
	b = MustValue(os.ReadFile("testdata/start/migration_success_completed.json"))
	MustValue(statusMap.Put(ctx, migrationID, serialization.JSON(b)))
	wg.Done()
}

func failureRunner(ctx context.Context, tcx it.TestContext, migrationID string, wg *sync.WaitGroup) {
	mSQL := fmt.Sprintf(`CREATE MAPPING IF NOT EXISTS %s TYPE IMap OPTIONS('keyFormat'='varchar', 'valueFormat'='json')`, migration.StatusMapName)
	MustValue(tcx.Client.SQL().Execute(ctx, mSQL))
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := MustValue(os.ReadFile("testdata/start/migration_success_initial.json"))
	MustValue(statusMap.Put(ctx, migrationID, serialization.JSON(b)))
	b = MustValue(os.ReadFile("testdata/start/migration_success_failure.json"))
	MustValue(statusMap.Put(ctx, migrationID, serialization.JSON(b)))
	wg.Done()
}

func findMigrationID(ctx context.Context, tcx it.TestContext, c chan string) {
	q := MustValue(tcx.Client.GetQueue(ctx, migration.StartQueueName))
	var b migration.ConfigBundle
	for {
		v := MustValue(q.PollWithTimeout(ctx, 2*time.Second))
		if v != nil {
			Must(json.Unmarshal(v.(serialization.JSON), &b))
			c <- b.MigrationID
			break
		}
	}
}

func fileExists(filename string) bool {
	MustValue(os.Getwd())
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
