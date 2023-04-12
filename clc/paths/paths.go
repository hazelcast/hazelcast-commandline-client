package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	EnvCLCHome    = "CLC_HOME"
	DefaultConfig = "config.yaml"
)

type FilterFn func(basePath string, entry os.DirEntry) (ok bool)

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
	if p := nearbyConfigPath(); p != "" {
		return p
	}
	p := filepath.Join(ResolveConfigDir("default"), DefaultConfig)
	if Exists(p) {
		return p
	}
	return ""
}

func DefaultLogPath(now time.Time) string {
	fn := fmt.Sprintf("%s.log", now.Format("2006-01-02"))
	return filepath.Join(Logs(), fn)
}

func ResolveConfigDir(name string) string {
	return Join(Configs(), name)
}

/*
ResolveConfigPath returns the normalized configuration path
The user has several ways of specifying the configuration:
 1. No configuration specified and there's config.yaml next to clc binary: config.yaml is used
 2. No configuration specified: The default configuration at $CLC_HOME/config.yaml is used
 3. Absolute path
 4. Relative path $PATH: $PATH in current working directory
 5. Relative path $PATH without extension: The configuration file at $CLC_HOME/$PATH/config.yaml is used.
*/
func ResolveConfigPath(path string) string {
	if path == "" {
		path = DefaultConfigPath()
	}
	if path == "" {
		return path
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

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func FindAll(cd string, fn FilterFn) ([]string, error) {
	var cs []string
	es, err := os.ReadDir(cd)
	if err != nil {
		return nil, err
	}
	for _, e := range es {
		if strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		if !fn(cd, e) {
			continue
		}
		cs = append(cs, e.Name())
	}
	return cs, nil
}

func nearbyConfigPath() string {
	// check whether there is config.yaml in the current directory
	wd, err := os.Getwd()
	if err == nil {
		// if no error
		path := Join(wd, DefaultConfig)
		if Exists(path) {
			return path
		}
	}
	e, err := os.Executable()
	if err == nil {
		// if no error
		dir := filepath.Dir(e)
		// check whether config.yaml exists in this dir
		path := Join(dir, DefaultConfig)
		if Exists(path) {
			return path
		}
	}
	return ""
}
