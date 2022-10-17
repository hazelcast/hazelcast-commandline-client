package paths_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it/skip"
)

func TestHomeDir_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	os.Setenv("HOME", "/dev/shm")
	assert.Equal(t, "/dev/shm/.local/share/clc", paths.HomeDir())
}

func TestDefaultConfigPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	assert.Equal(t, "/dev/shm/.local/share/clc/config.yaml", paths.DefaultConfigPath())
}

func TestDefaultLogPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	now := time.Date(2020, 2, 1, 9, 0, 0, 0, time.UTC)
	assert.Equal(t, "/dev/shm/.local/share/clc/logs/2020-02-01.log", paths.DefaultLogPath(now))
}
