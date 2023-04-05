package base_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it/skip"
)

func TestHome(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "home", f: homeTest_Unix},
		{name: "homeWithEnv", f: homeWithEnvTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func homeTest_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	homeTester(t, nil, func(t *testing.T, ec *it.ExecContext) {
		output := ec.StdoutText()
		target := check.MustValue(os.UserHomeDir()) + "/.hazelcast\n"
		assert.Equal(t, target, output)
	})
}

// TODO: TestHome_Windows

func homeWithEnvTest(t *testing.T) {
	skip.If(t, "os = windows")
	it.WithEnv(paths.EnvCLCHome, "/home/foo/dir", func() {
		homeTester(t, nil, func(t *testing.T, ec *it.ExecContext) {
			output := ec.StdoutText()
			target := "/home/foo/dir\n"
			assert.Equal(t, target, output)
		})
	})
}

func homeTester(t *testing.T, args []string, f func(t *testing.T, ec *it.ExecContext)) {
	cmd := &commands.HomeCommand{}
	cc := &it.CommandContext{}
	require.NoError(t, cmd.Init(cc))
	ec := it.NewExecuteContext(args)
	require.NoError(t, cmd.Exec(context.Background(), ec))
	f(t, ec)
}
