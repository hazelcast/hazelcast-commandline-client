package versioncmd_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/versioncmd"
)

func TestVersion(t *testing.T) {
	cmd := versioncmd.New()
	var b bytes.Buffer
	cmd.SetOut(&b)
	err := cmd.Execute()
	require.NoError(t, err)
	out := b.String()
	t.Log(out)
	require.Contains(t, out, "Hazelcast Command Line Client Version")
	require.Contains(t, out, "Latest Git Commit Hash")
	require.Contains(t, out, "Hazelcast Go Client Version")
	require.Contains(t, out, "Go Version")
}
