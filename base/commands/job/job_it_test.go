package job_test

import (
	"context"
	"fmt"
	"testing"

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
		{name: "listInteractive", f: listInteractiveTest},
		{name: "restartInteractive", f: restartInteractiveTest},
		{name: "restartNonInteractive", f: restartNonInteractiveTest},
		{name: "submitNonInteractive", f: submitNonInteractiveTest},
		{name: "submitInteractive", f: submitInteractiveTest},
		{name: "suspendResumeNonInteractive", f: suspendResumeNonInteractiveTest},
		{name: "suspendResumeInteractive", f: suspendResumeInteractiveTest},
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
				tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
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
				tcx.WriteStdinf("\\job submit --name %s %s --retries 0 --wait\n", name, jobPath)
				defer func() {
					tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
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
			tcx.AssertStdoutContains(name + "\tRUNNING")
		})
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
			tcx.CLCExecute(ctx, "job", "list")
			tcx.AssertStdoutNotContains(name)
		})
	})
}

func listInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		name := it.NewUniqueObjectName("job")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "submit", "--name", name, jobPath, "--retries", "0", "--wait")
				tcx.AssertStderrContains("OK")
			})
			tcx.WithReset(func() {
				tcx.WriteStdinString("\\job list\n")
				tcx.AssertStdoutDollar(fmt.Sprintf("%s$|$RUNNING", name))
			})
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
				tcx.WriteStdinString("\\job list\n")
				tcx.AssertStdoutNotContains(name)
			})
		})
	})
}

func suspendResumeNonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		name := it.NewUniqueObjectName("job")
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "submit", "--name", name, jobPath, "--wait")
			tcx.AssertStderrContains("OK")
		})
		defer func() {
			tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
		}()
		tcx.WithReset(func() {
			tcx.WriteStdinf("\\job suspend %s --wait\n", name)
			tcx.AssertStderrContains("OK")
		})
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "list")
			tcx.AssertStdoutContains(name + "\tSUSPENDED")
		})
		tcx.WithReset(func() {
			tcx.WriteStdinf("\\job resume %s --wait\n", name)
			tcx.AssertStderrContains("OK")
		})
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "list")
			tcx.AssertStdoutContains(name + "\tRUNNING")
		})
	})
}

func suspendResumeInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		name := it.NewUniqueObjectName("job")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "submit", "--name", name, jobPath, "--wait")
				tcx.AssertStderrContains("OK")
			})
			defer func() {
				tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
			}()
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "suspend", name, "--wait")
				tcx.AssertStderrContains("OK")
			})
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "list")
				tcx.AssertStdoutContains(name + "\tSUSPENDED")
			})
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "resume", name, "--wait")
				tcx.AssertStderrContains("OK")
			})
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "list")
				tcx.AssertStdoutContains(name + "\tRUNNING")
			})
		})
	})
}

func restartNonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		name := it.NewUniqueObjectName("job")
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "submit", "--name", name, jobPath, "--wait")
			tcx.AssertStderrContains("OK")
		})
		defer func() {
			tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
		}()
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "restart", name, "--wait")
			tcx.AssertStderrContains("OK")
		})
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "job", "list")
			tcx.AssertStdoutContains(name + "\tRUNNING")
		})
	})
}

func restartInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		name := it.NewUniqueObjectName("job")
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "submit", "--name", name, jobPath, "--wait")
				tcx.AssertStderrContains("OK")
			})
			defer func() {
				tcx.CLCExecute(ctx, "job", "cancel", name, "--wait")
			}()
			tcx.WithReset(func() {
				tcx.WriteStdinf("\\job restart %s --wait\n", name)
				tcx.AssertStderrContains("OK")
			})
			tcx.WithReset(func() {
				tcx.CLCExecute(ctx, "job", "list")
				tcx.AssertStdoutContains(name + "\tRUNNING")
			})
		})
	})
}
