//go:build std || topic

package topic_test

import (
	"context"
	"sync"
	"testing"
	"time"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestTopic(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Publish_NonInteractive", f: publish_NonInteractiveTest},
		{name: "Subscribe_NonInteractive", f: subscribe_NonInteractiveTest},
		{name: "Subscribe_Cancel_NonInteractiveTest", f: subscribe_Cancel_NonInteractiveTest},
		{name: "Destroy_NonInteractive", f: destroy_NonInteractiveTest},
		{name: "Destroy_AutoYes_NonInteractiveTest", f: destroy_autoYes_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func publish_NonInteractiveTest(t *testing.T) {
	it.TopicTester(t, func(tcx it.TestContext, tp *hz.Topic) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			var values []string
			valuesMu := &sync.Mutex{}
			sid := check.MustValue(tp.AddMessageListener(ctx, func(event *hz.MessagePublished) {
				valuesMu.Lock()
				values = append(values, event.Value.(string))
				valuesMu.Unlock()
			}))
			defer func() { check.Must(tp.RemoveListener(ctx, sid)) }()
			check.Must(tcx.CLC().Execute(ctx, "topic", "-n", tp.Name(), "publish", "value1"))
			check.Must(tcx.CLC().Execute(ctx, "topic", "-n", tp.Name(), "publish", "value2"))
			require.Eventually(t, func() bool {
				valuesMu.Lock()
				ok := slices.Contains(values, "value1")
				valuesMu.Unlock()
				return ok
			}, 5*time.Second, 100*time.Millisecond)
			require.Eventually(t, func() bool {
				valuesMu.Lock()
				ok := slices.Contains(values, "value2")
				valuesMu.Unlock()
				return ok
			}, 5*time.Second, 100*time.Millisecond)
		})
	})
}

func subscribe_NonInteractiveTest(t *testing.T) {
	it.TopicTester(t, func(tcx it.TestContext, tp *hz.Topic) {
		ctx := context.Background()
		tcx.WithReset(func() {
			go func() {
				check.Must(tcx.CLC().Execute(ctx, "topic", "-n", tp.Name(), "subscribe", "--count", "2"))
			}()
			time.Sleep(1 * time.Second)
			check.Must(tp.PublishAll(ctx, "value1", "value2"))
			tcx.AssertStdoutContains("value1")
			tcx.AssertStdoutContains("value2")
		})
	})
}

func subscribe_Cancel_NonInteractiveTest(t *testing.T) {
	t.Skipf("Disabling this test, since it requires some internal changes in CLC")
	it.TopicTester(t, func(tcx it.TestContext, tp *hz.Topic) {
		ctx := context.Background()
		tcx.WithReset(func() {
			ctx, cancel := context.WithCancel(ctx)
			go func() {
				check.Must(tcx.CLC().Execute(ctx, "topic", "-n", tp.Name(), "subscribe"))
			}()
			time.Sleep(1 * time.Second)
			check.Must(tp.PublishAll(ctx, "value1", "value2"))
			tcx.AssertStdoutContains("value1")
			tcx.AssertStdoutContains("value2")
			cancel()
			tcx.AssertStderrContains("OK")
		})
	})
}

func destroy_NonInteractiveTest(t *testing.T) {
	it.TopicTester(t, func(tcx it.TestContext, tp *hz.Topic) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			go tcx.WriteStdin([]byte("y\n"))
			check.Must(tcx.CLC().Execute(ctx, "topic", "-n", tp.Name(), "destroy"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameTopic, tp.Name(), objects))
		})
	})
}

func destroy_autoYes_NonInteractiveTest(t *testing.T) {
	it.TopicTester(t, func(tcx it.TestContext, tp *hz.Topic) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "topic", "-n", tp.Name(), "destroy", "--yes"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameTopic, tp.Name(), objects))
		})
	})
}

func objectExists(sn, name string, objects []types.DistributedObjectInfo) bool {
	for _, obj := range objects {
		if sn == obj.ServiceName && name == obj.Name {
			return true
		}
	}
	return false
}
