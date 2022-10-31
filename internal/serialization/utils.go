package serialization

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

const (
	PortablePathSuffix = ".portable.json"
	CompactPathSuffix  = ".compact.json"
)

type recurseCallback func(path string)

func LoadStructFromJSON[T any](b []byte) (T, error) {
	var gp T
	if err := json.Unmarshal(b, &gp); err != nil {
		return gp, err
	}
	return gp, nil
}

func LoadStructsFromPaths[T any](suffix string, paths ...string) ([]T, error) {
	ps := make([]T, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(path, ".") {
			continue
		}
		if !strings.HasSuffix(path, suffix) {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading path: %s: %w", path, err)
		}
		p, err := LoadStructFromJSON[T](b)
		if err != nil {
			return nil, fmt.Errorf("error loading from path: %s: %w", path, err)
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func UpdateSerializationConfigWithRecursivePaths(cfg *hazelcast.Config, lg log.Logger, paths ...string) error {
	var portablePaths []string
	var compactPaths []string
	cb := func(path string) {
		if strings.HasSuffix(path, PortablePathSuffix) {
			portablePaths = append(portablePaths, path)
		} else if strings.HasSuffix(path, CompactPathSuffix) {
			compactPaths = append(compactPaths, path)
		}
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			lg.Warn("Checking schemas directory: %s", err.Error())
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
	// portable stuff
	gps, err := LoadStructsFromPaths[GenericPortable](PortablePathSuffix, portablePaths...)
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
	// compact stuff
	gcs, err := LoadStructsFromPaths[GenericCompact](CompactPathSuffix, compactPaths...)
	if err != nil {
		return err
	}
	gs := make([]serialization.CompactSerializer, len(gcs))
	for i, f := range gcs {
		cs, err := NewGenericCompact(f)
		if err != nil {
			lg.Warn("Could not create GenericCompact for type: %s: %s", f.ValueTypeName, err.Error())
		}
		gs[i] = cs
	}
	cfg.Serialization.Compact.SetSerializers(gs...)
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
