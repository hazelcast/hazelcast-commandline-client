package paths

import (
	"path/filepath"
)

func homeDir(userHomeDir string) string {
	return filepath.Join(userHomeDir, "AppData/Roaming/Hazelcast")
}
