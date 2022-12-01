package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func TestConfigDirFile(t *testing.T) {
	// ignoring the error
	_ = os.Setenv(paths.EnvCLCHome, "/home/clc")
	defer os.Unsetenv(paths.EnvCLCHome)
	td := check.MustValue(os.MkdirTemp("", "clctest-*"))
	existingDir := filepath.Join(td, "mydir")
	check.Must(os.MkdirAll(existingDir, 0700))
	existingFile := filepath.Join(td, "mydir", "myconfig.yaml")
	check.Must(os.WriteFile(existingFile, []byte{}, 0700))
	testCases := []struct {
		descr       string
		cfgPath     string
		defaultPath string
		targetDir   string
		targetFile  string
	}{
		// if cfgPath is given, defaultPath shouldn't be used
		{
			descr:       "if cfgPath is given, defaultPath shouldn't be used",
			cfgPath:     "my-config",
			defaultPath: "default",
			targetDir:   "/home/clc/configs/my-config",
			targetFile:  "config.yaml",
		},
		{
			descr:       "if cfgPath is not given, defaultPath should be used",
			defaultPath: "default-cfg",
			targetDir:   "/home/clc/configs/default-cfg",
			targetFile:  "config.yaml",
		},
		{
			descr:      "existing cfg dir",
			cfgPath:    existingDir,
			targetDir:  existingDir,
			targetFile: "config.yaml",
		},
		{
			descr:      "existing cfg file",
			cfgPath:    existingFile,
			targetDir:  existingDir,
			targetFile: "myconfig.yaml",
		},
		{
			descr:      "nonexistent dir",
			cfgPath:    "/home/me/foo",
			targetDir:  "/home/me/foo",
			targetFile: "config.yaml",
		},
		{
			descr:      "nonexistent file",
			cfgPath:    "/home/me/foo/some.file",
			targetDir:  "/home/me/foo",
			targetFile: "some.file",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.descr, func(t *testing.T) {
			d, f, err := configDirFile(tc.cfgPath, tc.defaultPath)
			assert.NoError(t, err)
			assert.Equal(t, tc.targetDir, d)
			assert.Equal(t, tc.targetFile, f)
		})
	}
}
