package serialization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

type recurseCallback func(path string)

func UpdateSerializationConfigWithRecursivePaths(cfg *hazelcast.Config, lg log.Logger, paths ...string) error {
	var portablePaths []string
	cb := func(path string) {
		if strings.HasSuffix(path, PortablePathSuffix) {
			portablePaths = append(portablePaths, path)
		}
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			lg.Warn("Cannot stat %s: %s", path, err.Error())
			continue
		}
		if info.IsDir() {
			if err := recurseDirectory(path, cb, lg); err != nil {
				lg.Warn("Cannot traverse directory %s: %s", path, err.Error())
			}
			continue
		}
		cb(path)
	}
	gps, err := LoadPortablesFromPaths(portablePaths...)
	if err != nil {
		return err
	}
	gpfs, err := NewPortableFactoriesFromItems(gps...)
	if err != nil {
		return err
	}
	fs := make([]serialization.PortableFactory, len(gpfs))
	for i, f := range gpfs {
		fs[i] = f
	}
	cfg.Serialization.SetPortableFactories(fs...)
	return nil
}

func recurseDirectory(dir string, cb recurseCallback, lg log.Logger) error {
	es, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading dir: %s: %w", dir, err)
	}
	var dirs []string
	for _, e := range es {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			dirs = append(dirs, path)
			continue
		}
		cb(path)
	}
	for _, dir = range dirs {
		// ignoring the error while traversing directories
		if err := recurseDirectory(dir, cb, lg); err != nil {
			lg.Warn("Cannot traverse directory %s: %s", dir, err.Error())
		}
	}
	return nil
}
