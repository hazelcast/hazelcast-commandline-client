package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	unixDir       = ".local/share/clc"
	EnvCLCHome    = "CLC_HOME"
	DefaultConfig = "config.yaml"
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

func ConfigsDir() string {
	return filepath.Join(HomeDir(), "configs")
}

func SchemasDir() string {
	return filepath.Join(HomeDir(), "schemas")
}

func LogsDir() string {
	return filepath.Join(HomeDir(), "logs")
}

func DefaultConfigPath() string {
	return filepath.Join(HomeDir(), "config.yaml")
}

func DefaultLogPath(now time.Time) string {
	fn := fmt.Sprintf("%s.log", now.Format("2006-01-02"))
	return filepath.Join(LogsDir(), fn)
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
		path = filepath.Join(ConfigsDir(), path, DefaultConfig)
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
