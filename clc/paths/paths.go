package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	unixDir    = ".local/share/clc"
	EnvCLCHome = "CLC_HOME"
)

func HomeDir() string {
	dir := os.Getenv(EnvCLCHome)
	if dir == "" {
		cd, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		dir = filepath.Join(cd, unixDir)
	}
	return dir
}

func DefaultConfigPath() string {
	return filepath.Join(HomeDir(), "config.yaml")
}

func DefaultLogPath(now time.Time) string {
	fn := fmt.Sprintf("%s.log", now.Format("2006-01-02"))
	return filepath.Join(HomeDir(), "logs", fn)
}
