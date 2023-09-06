//go:build migration

package migration_test

import (
	"context"
	"encoding/json"
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
		go tcx.WithReset(func() {
			Must(tcx.CLC().Execute(ctx, "start", "dmt-config", "--yes"))
		})
		successfulRunner(tcx, ctx)
		for _, m := range []string{"first message", "second message", "last message", "status report"} {
			tcx.AssertStdoutContains(m)
		}
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
		for _, m := range []string{"first message", "second message", "fail status report"} {
			tcx.AssertStdoutContains(m)
		}
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
