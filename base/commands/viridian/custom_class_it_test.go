package viridian_test

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"strings"
	"testing"
	"time"
)

func customClassTest_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		customClassTest(ctx, tcx)
	})
}

func customClass_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			customClassTest(ctx, tcx)
		})
	})
}

func customClassTest(ctx context.Context, tcx it.TestContext) {
	// setup
	f := "foo.zip"
	fd := "testdata/" + f
	cid := ensureClusterRunning(ctx, tcx)
	// test upload custom class
	tcx.WithReset(func() {
		tcx.CLCExecute(ctx, "viridian", "upload-custom-class", cid, fd)
		tcx.AssertStderrContains("OK")
	})
	tcx.WithReset(func() {
		check.Must(waitCustomClassUpload(ctx, tcx))
		tcx.AssertStderrContains("OK")
	})
	id := ""
	// test list custom class
	tcx.WithReset(func() {
		tcx.CLCExecute(ctx, "viridian", "list-custom-classes", cid)
		tcx.AssertStderrContains("OK")
		id = customClassID(tcx.ExpectStdout.String())
		tcx.AssertStdoutContains(f)
	})
	// test download custom class
	tcx.WithReset(func() {
		tcx.CLCExecute(ctx, "viridian", "download-custom-class", cid, f)
		tcx.AssertStderrContains("OK")
		tcx.AssertStderrContains("Custom class downloaded successfully.")
	})
	// test delete custom class
	tcx.WithReset(func() {
		check.Must(waitState(ctx, tcx, cid, "RUNNING"))
		tcx.CLCExecute(ctx, "viridian", "delete-custom-class", cid, id)
		tcx.AssertStderrContains("OK")
	})
	tcx.WithReset(func() {
		tcx.CLCExecute(ctx, "viridian", "list-custom-classes", cid)
		tcx.AssertStderrContains("OK")
		tcx.AssertStdoutNotContains(f)
	})
}

func waitCustomClassUpload(ctx context.Context, tcx it.TestContext) error {
	tryCount := 0
	for {
		if tryCount == 5 {
			return fmt.Errorf("custom class uploaded exceeded try limit")
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if strings.Contains(tcx.ExpectStdout.String(), "Custom class uploaded successfully.") {
			return nil
		}
		tryCount++
		time.Sleep(5 * time.Second)
	}
}

func customClassID(l string) string {
	return strings.Split(l, "\t")[0]
}
