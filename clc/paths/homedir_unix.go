//go:build !windows

package paths

import (
	"path/filepath"
)

func homeDir(userHomeDir string) string {
	return filepath.Join(userHomeDir, ".local/share/clc")
}
