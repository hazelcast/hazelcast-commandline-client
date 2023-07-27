//go:build std || shell

package commands_test

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/object"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it/skip"
)

func TestShell(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "ShellErrors", f: shellErrorsTest},
		{name: "ShellNoDoubleError", f: shellNoDoubleErrorTest},
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

func shellNoDoubleErrorTest(t *testing.T) {
	// this test times out on Windows on CI,
	// but passes on Windows on local.
	// so skipping for now... --YT
	skip.If(t, "os = windows")
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			for _, text := range []string{"foo;", "\\foo", "\\map --foo"} {
				tcx.WithReset(func() {
					tcx.WriteStdinString(text + "\n")
					tcx.AssertStderrNotRegexMatch("Error:.*\nError:")
				})
			}
		})
	})
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
