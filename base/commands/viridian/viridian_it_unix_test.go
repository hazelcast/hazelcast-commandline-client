//go:build unix && (std || viridian)

package viridian_test

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func streamLogs_nonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
		go func() {
			time.Sleep(10 * time.Second)
			t.Logf("Sending interrupt signal to this process")
			// ignoring the error here
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()
		tcx.CLCExecute(ctx, "viridian", "stream-logs", c.ID)
		tcx.AssertStdoutContains("Loading configuration")
	})
}
