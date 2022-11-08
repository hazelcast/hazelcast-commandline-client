package shell_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
)

func TestIsPipe(t *testing.T) {
	// since the test is not run in a pipe, it will always return false.
	assert.False(t, shell.IsPipe())
}
