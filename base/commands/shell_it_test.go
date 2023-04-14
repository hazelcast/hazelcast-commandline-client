package commands_test

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/object"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestShell(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "ShellErrors", f: shellErrorsTest},
		{name: "ShellHelp", f: shellHelpTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func shellErrorsTest(t *testing.T) {
	testCases := []struct {
		name    string
		command string
		errText string
	}{
		{
			name:    "invalid command",
			command: "\\foobar",
			errText: "unknown command \\foobar",
		},
		{
			name:    "invalid flag",
			command: "\\object list --foobar",
			errText: "unknown flag: --foobar",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				ctx := context.Background()
				tcx.WithShell(ctx, func(tcx it.TestContext) {
					tcx.WithReset(func() {
						tcx.WriteStdinString(tc.command + "\n")
						tcx.AssertStderrEquals(fmt.Sprintf("Error: %s\n", tc.errText))
					})
				})
			})
		})
	}
}

func shellHelpTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinString("\\help\n")
				tcx.AssertStdoutContains("Usage:")
			})
		})
	})
}
