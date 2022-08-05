package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const expectedMsg = `The support for List hasn't been implemented yet.

If you would like us to implement it, please drop by at:
https://github.com/hazelcast/hazelcast-commandline-client/issues/48 and add a thumbs up 👍.
We're happy to implement it quickly based on demand!
`

func TestNewFakeCommand(t *testing.T) {
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
