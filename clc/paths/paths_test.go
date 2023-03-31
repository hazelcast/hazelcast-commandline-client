package paths_test

import (
	"fmt"
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
	assert.Equal(t, "/dev/shm/.hazelcast", paths.Home())
}

func TestHomeDir_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast`, paths.Home())
}

func TestDefaultConfigPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	assert.Equal(t, "", paths.DefaultConfigPath())
}

func TestDefaultConfigPath_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast\configs\default\config.yaml`, paths.DefaultConfigPath())
}

func TestDefaultLogPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	now := time.Date(2020, 2, 1, 9, 0, 0, 0, time.UTC)
	assert.Equal(t, "/dev/shm/.hazelcast/logs/2020-02-01.log", paths.DefaultLogPath(now))
}

func TestDefaultLogPath_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	now := time.Date(2020, 2, 1, 9, 0, 0, 0, time.UTC)
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast\logs\2020-02-01.log`, paths.DefaultLogPath(now))
}

func TestSchemas_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	assert.Equal(t, "/dev/shm/.hazelcast/schemas", paths.Schemas())
}

func TestSchemas_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast\schemas`, paths.Schemas())
}

func TestResolveConfigPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	// default config
	assert.Equal(t, "", paths.ResolveConfigPath(""))
	// path to the configuration file
	assert.Equal(t, "/etc/hz.yaml", paths.ResolveConfigPath("/etc/hz.yaml"))
	// configuration name
	assert.Equal(t, "/dev/shm/.hazelcast/configs/pr-3066/config.yaml", paths.ResolveConfigPath("pr-3066"))
}

func TestResolveConfigPath_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	// default config
	assert.Equal(t, "", paths.ResolveConfigPath(""))
	// path to the configuration file
	assert.Equal(t, `C:\hz.yaml`, paths.ResolveConfigPath(`C:\hz.yaml`))
	// configuration name
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	assert.Equal(t, `C:\Users\foo\AppData\Roaming\Hazelcast\configs\pr-3066\config.yaml`, paths.ResolveConfigPath("pr-3066"))
}

func TestResolveLogPath_Unix(t *testing.T) {
	skip.If(t, "os = windows")
	Must(os.Setenv("HOME", "/dev/shm"))
	now := time.Now()
	path := fmt.Sprintf("/dev/shm/.hazelcast/logs/%s.log", now.Format("2006-01-02"))
	assert.Equal(t, path, paths.ResolveLogPath(""))
	assert.Equal(t, "/var/hz.log", paths.ResolveLogPath("/var/hz.log"))
}

func TestResolveLogPath_Windows(t *testing.T) {
	skip.IfNot(t, "os = windows")
	Must(os.Setenv("USERPROFILE", `C:\Users\foo`))
	now := time.Now()
	path := fmt.Sprintf(`C:\Users\foo\AppData\Roaming\Hazelcast\logs\%s.log`, now.Format("2006-01-02"))
	assert.Equal(t, path, paths.ResolveLogPath(""))
	assert.Equal(t, `C:\hz.log`, paths.ResolveLogPath(`C:\hz.log`))
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
