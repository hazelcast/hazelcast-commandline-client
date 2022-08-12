package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFakeCommand(t *testing.T) {
	const expectedMsg = `The support for List isn't implemented yet.
Add a thumbs up to it at: https://github.com/hazelcast/hazelcast-commandline-client/issues/48
`
	fd := FakeDoor{Name: "List", IssueNum: 48}
	cmd := NewFakeCommand(fd)
	ctx := context.Background()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	_, err := cmd.ExecuteContextC(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedMsg, stdout.String())
	require.Empty(t, stderr.String())
}
