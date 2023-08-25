//go:build std || version

package commands_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestVersion(t *testing.T) {
	internal.Version = "v5.2.0"
	cmd := &commands.VersionCommand{}
	cc := &it.CommandContext{}
	require.NoError(t, cmd.Init(cc))
	ec := it.NewExecuteContext(nil)
	ec.Set(clc.PropertyFormat, base.PrinterDelimited)
	require.NoError(t, cmd.Exec(context.TODO(), ec))
	assert.Equal(t, "v5.2.0\n", ec.StdoutText())
	ec.Set(clc.PropertyVerbose, true)
	require.NoError(t, cmd.Exec(context.TODO(), ec))
	assert.Equal(t, ec.Rows[0][0].Value, "Hazelcast CLC")
	assert.Contains(t, ec.Rows[1][0].Value, "Latest Git Commit Hash")
	assert.Contains(t, ec.Rows[2][0].Value, "Hazelcast Go Client")
	assert.Contains(t, ec.Rows[3][0].Value, "Go")
}
