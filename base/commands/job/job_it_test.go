package job_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

const jobPath = "testdata/sample-job-1-1.0-SNAPSHOT-all.jar"

func TestJob(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "listNonInteractive", f: listNonInteractiveTest},
		{name: "submitNonInteractive", f: submitNonInteractiveTest},
		{name: "submitInteractive", f: submitInteractiveTest},
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
			tcx.CLCExecute(ctx, "job", "submit", "--name", name, jobPath, "--retries", "0", "--wait")
			defer func() {
				ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
				defer cancel()
				tcx.CLCExecute(ctx, "job", "cancel", name)
			}()
			tcx.AssertStderrContains("OK")
		})
	})
}

func submitInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				name := it.NewUniqueObjectName("job")
				c := fmt.Sprintf("\\job submit --name %s %s --retries 0 --wait\n", name, jobPath)
				tcx.WriteStdin([]byte(c))
				defer func() {
					ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
					defer cancel()
					tcx.CLCExecute(ctx, "job", "cancel", name)
				}()
				tcx.AssertStderrContains("OK")
			})
		})
	})
}

func listNonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		name := it.NewUniqueObjectName("job")
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "submit", "--name", name, jobPath, "--retries", "0", "--wait")
			tcx.AssertStderrContains("OK")
		})
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "list")
			tcx.AssertStdoutContains(name)
		})
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
			tcx.CLCExecute(ctx, "job", "list")
			tcx.AssertStdoutNotContains(name)
		})
	})
}
