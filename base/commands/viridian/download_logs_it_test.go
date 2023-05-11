package viridian_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func downloadLogs_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		dir := check.MustValue(os.MkdirTemp("", "log"))
		defer func() { check.Must(os.RemoveAll(dir)) }()
		c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "viridian", "download-logs", c.ID, "--output-dir", dir)
			tcx.AssertStderrContains("OK")
			require.FileExists(t, paths.Join(dir, "node-1.log"))
			require.FileExists(t, paths.Join(dir, "node-2.log"))
			require.FileExists(t, paths.Join(dir, "node-3.log"))
		})
	})
}

func downloadLogs_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		dir := check.MustValue(os.MkdirTemp("", "log"))
		defer func() { check.Must(os.RemoveAll(dir)) }()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				c := createOrGetClusterWithState(ctx, tcx, "")
				tcx.WriteStdinf("\\viridian download-logs %s -o %s\n", c.Name, dir)
				tcx.AssertStderrContains("OK")
				require.FileExists(t, paths.Join(dir, "node-1.log"))
				require.FileExists(t, paths.Join(dir, "node-2.log"))
				require.FileExists(t, paths.Join(dir, "node-3.log"))
			})
		})
	})
}
