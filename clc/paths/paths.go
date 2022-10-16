package paths

import (
	"os"
	"path/filepath"
)

const (
	unixDir    = ".local/share/hz-cli"
	envCLCHome = "CLC_HOME"
)

func HomeDir() string {
	dir := os.Getenv(envCLCHome)
	if dir == "" {
		cd, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		dir = filepath.Join(cd, unixDir)
	}
	return dir
}
