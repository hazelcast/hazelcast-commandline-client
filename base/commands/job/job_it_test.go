package job_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestJob(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "submitNonInteractive", f: submitNonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func submitNonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithReset(func() {
			name := it.NewUniqueObjectName("job")
			tcx.CLCExecute(ctx, "job", "submit", "--name", name, "testdata/sample-job-1-1.0-SNAPSHOT-all.jar")
			defer func() {
				ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
				defer cancel()
				//tcx.CLC().
				//check.Must(job.WaitJobState(ctx, tcx.)
				tcx.CLCExecute(ctx, "job", "cancel", name)
			}()
			tcx.AssertStderrContains("OK")
		})
	})
}
