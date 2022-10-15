package commands_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestVersion(t *testing.T) {
	cmd := &commands.VersionCommand{}
	require.NoError(t, cmd.Init(&it.CommandContext{}))
	ec := it.NewExecuteContext(nil)
	require.NoError(t, cmd.Exec(ec))
	output := ec.StdoutText()
	t.Log(output)
	require.Contains(t, output, "Hazelcast Command Line Client Version")
	require.Contains(t, output, "Latest Git Commit Hash")
	require.Contains(t, output, "Hazelcast Go Client Version")
	require.Contains(t, output, "Go Version")
}
