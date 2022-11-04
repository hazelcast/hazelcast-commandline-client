package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	EnvCLCHome    = "CLC_HOME"
	DefaultConfig = "config.yaml"
)

func Home() string {
	dir := os.Getenv(EnvCLCHome)
	if dir != "" {
		return dir
	}
	d, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDir(d)
}

func Configs() string {
	return filepath.Join(Home(), "configs")
}

func Schemas() string {
	return filepath.Join(Home(), "schemas")
}

func Logs() string {
	return filepath.Join(Home(), "logs")
}

func DefaultConfigPath() string {
	return filepath.Join(Configs(), "default", "config.yaml")
}

func DefaultLogPath(now time.Time) string {
	fn := fmt.Sprintf("%s.log", now.Format("2006-01-02"))
	return filepath.Join(Logs(), fn)
}

/*
ResolveConfigPath returns the normalized configuration path
The user has several ways of specifying the configuration:
 1. No configuration: The default configuration at $CLC_HOME/config.yaml is used
 2. Absolute path
 3. Relative path $PATH: $PATH in current working directory
 4. Relative path $PATH without extension: The configuration file at $CLC_HOME/$PATH/config.yaml is used.
*/
func ResolveConfigPath(path string) string {
	if path == "" {
		path = DefaultConfigPath()
	}
	if filepath.Ext(path) == "" {
		path = filepath.Join(Configs(), path, DefaultConfig)
	}
	return path
}

func ResolveLogPath(path string) string {
	if path == "" {
		return DefaultLogPath(time.Now())
	}
	return path
}

func Join(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}
	path := paths[0]
	for _, p := range paths[1:] {
		if filepath.IsAbs(p) {
			path = p
			continue
		}
		path = filepath.Join(path, p)
	}
	return path
}
