//go:build std || migration

package migration_test

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands/migration"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

func TestEstimate(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		createMapping(ctx, tcx)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			check.Must(tcx.CLC().Execute(ctx, "estimate", "testdata/dmt-config"))
		}()
		c := make(chan string)
		go findEstimationID(ctx, tcx, c)
		mID := <-c
		go estimateRunner(ctx, tcx, mID)
		wg.Wait()
		tcx.AssertStdoutContains("OK Estimation completed successfully")
	})
}

func estimateRunner(ctx context.Context, tcx it.TestContext, migrationID string) {
	statusMap := check.MustValue(tcx.Client.GetMap(ctx, migration.StatusMapName))
	b := check.MustValue(os.ReadFile("testdata/estimate/estimate_completed.json"))
	check.Must(statusMap.Set(ctx, migrationID, serialization.JSON(b)))
}

func findEstimationID(ctx context.Context, tcx it.TestContext, c chan string) {
	q := check.MustValue(tcx.Client.GetQueue(ctx, migration.EstimateQueueName))
	var b migration.ConfigBundle
	for {
		v := check.MustValue(q.PollWithTimeout(ctx, 100*time.Millisecond))
		if v != nil {
			check.Must(json.Unmarshal(v.(serialization.JSON), &b))
			c <- b.MigrationID
			break
		}
	}
}
