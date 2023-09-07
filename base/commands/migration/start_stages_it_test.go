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
)

func TestMigration(t *testing.T) {
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
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		var wg sync.WaitGroup
		wg.Add(1)
		go tcx.WithReset(func() {
			defer wg.Done()
			Must(tcx.CLC().Execute(ctx, "start", "dmt-config", "--yes"))
		})
		successfulRunner(tcx, ctx)
		tcx.AssertStdoutContains(`
Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.
	
Selected data structures in the source cluster will be migrated to the target cluster.	


 OK   [1/3] Connected to the migration cluster.
first message
 OK   [2/3] Started the migration.
second message
last message
status report
 OK   [3/3] Migrated the cluster.

 OK   Migration completed successfully.`)
	})
}

func startTest_Failure(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		go tcx.WithReset(func() {
			tcx.CLC().Execute(ctx, "start", "dmt-config", "--yes")
		})
		failureRunner(tcx, ctx)
		tcx.AssertStdoutContains(`
Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.
	
Selected data structures in the source cluster will be migrated to the target cluster.	


 OK   [1/3] Connected to the migration cluster.
first message
 OK   [2/3] Started the migration.
second message
fail status report
 FAIL Could not migrate the cluster: migration failed with following error(s):
error1
error2`)
	})
}

func successfulRunner(tcx it.TestContext, ctx context.Context) {
	c := make(chan string, 1)
	go findMigrationID(ctx, tcx, c)
	migrationID := <-c
	topic := MustValue(tcx.Client.GetTopic(ctx, migration.MakeUpdateTopicName(migrationID)))
	Must(topic.Publish(ctx, migration.UpdateMessage{Status: migration.StatusInProgress, Message: "first message"}))
	Must(topic.Publish(ctx, migration.UpdateMessage{Status: migration.StatusInProgress, Message: "second message"}))
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.MakeStatusMapName(migrationID)))
	b := MustValue(json.Marshal(migration.MigrationStatus{
		Status: migration.StatusComplete,
		Report: "status report",
		Logs:   []string{"log1", "log2"},
	}))
	Must(statusMap.Set(ctx, migration.StatusMapEntryName, serialization.JSON(b)))
	Must(topic.Publish(ctx, migration.UpdateMessage{Status: migration.StatusComplete, Message: "last message"}))
}

func failureRunner(tcx it.TestContext, ctx context.Context) {
	c := make(chan string, 1)
	go findMigrationID(ctx, tcx, c)
	migrationID := <-c
	topic := MustValue(tcx.Client.GetTopic(ctx, migration.MakeUpdateTopicName(migrationID)))
	Must(topic.Publish(ctx, migration.UpdateMessage{Status: migration.StatusInProgress, Message: "first message"}))
	statusMap := MustValue(tcx.Client.GetMap(ctx, migration.MakeStatusMapName(migrationID)))
	b := MustValue(json.Marshal(migration.MigrationStatus{
		Status: migration.StatusFailed,
		Report: "fail status report",
		Errors: []string{"error1", "error2"},
	}))
	Must(statusMap.Set(ctx, migration.StatusMapEntryName, serialization.JSON(b)))
	Must(topic.Publish(ctx, migration.UpdateMessage{Status: migration.StatusFailed, Message: "second message"}))
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
