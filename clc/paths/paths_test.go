package paths_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it/skip"
)

func TestHomeDir_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	assert.Equal(t, "/dev/shm/.local/share/clc", paths.Home())
}

func TestHomeDir_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast CLC`, paths.Home())
}

func TestDefaultConfigPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	assert.Equal(t, "/dev/shm/.local/share/clc/configs/default/config.yaml", paths.DefaultConfigPath())
}

func TestDefaultConfigPath_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast CLC\configs\default\config.yaml`, paths.DefaultConfigPath())
}

func TestDefaultLogPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	now := time.Date(2020, 2, 1, 9, 0, 0, 0, time.UTC)
	assert.Equal(t, "/dev/shm/.local/share/clc/logs/2020-02-01.log", paths.DefaultLogPath(now))
}

func TestDefaultLogPath_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	now := time.Date(2020, 2, 1, 9, 0, 0, 0, time.UTC)
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast CLC\logs\2020-02-01.log`, paths.DefaultLogPath(now))
}

func TestJoin(t *testing.T) {
	skip.If(t, "os = windows")
	testCases := []struct {
		paths  []string
		result string
	}{
		{
			paths:  nil,
			result: "",
		},
		{
			paths:  []string{"foo"},
			result: "foo",
		},
		{
			paths:  []string{"foo", "bar"},
			result: "foo/bar",
		},
		{
			paths:  []string{"foo/bar", "zoo"},
			result: "foo/bar/zoo",
		},
		{
			paths:  []string{"foo/bar", "zoo", ""},
			result: "foo/bar/zoo",
		},
		{
			paths:  []string{"/foo/bar", "zoo", ""},
			result: "/foo/bar/zoo",
		},
		{
			paths:  []string{"/foo/bar", "/zoo", ""},
			result: "/zoo",
		},
		{
			paths:  []string{"/foo/bar", "/zoo", "baz"},
			result: "/zoo/baz",
		},
	}
	for _, tc := range testCases {
		t.Run(strings.Join(tc.paths, ":"), func(t *testing.T) {
			assert.Equal(t, tc.result, paths.Join(tc.paths...))
		})
	}
}
