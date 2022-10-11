package serialization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type recurseCallback func(path string)

func UpdateConfigWithRecursivePaths(cfg *hazelcast.Config, paths ...string) error {
	var portablePaths []string
	cb := func(path string) {
		if strings.HasSuffix(path, PortablePathSuffix) {
			portablePaths = append(portablePaths, path)
		}
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			// TODO: log the error
			continue
		}
		if info.IsDir() {
			if err := recurseDirectory(path, cb); err != nil {
				// TODO: log the error
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

func recurseDirectory(dir string, cb recurseCallback) error {
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
		// TODO: log the error
		_ = recurseDirectory(dir, cb)
	}
	return nil
}
