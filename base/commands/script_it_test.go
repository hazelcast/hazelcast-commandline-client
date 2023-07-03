package commands_test

import (
	"context"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestScript(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "script_NonInteractive", f: script_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func script_NonInteractiveTest(t *testing.T) {
	ctx := context.TODO()
	it.MapTester(t, func(tcx it.TestContext, m *hazelcast.Map) {
		tcx.CLCExecute(ctx, "script", "testdata/test-script.clc", "--echo", "--ignore-errors")
		tcx.AssertStdoutContains("bar")
		tcx.AssertStderrContains("unknown command")
	})
}
