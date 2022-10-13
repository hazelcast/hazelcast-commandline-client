package serialization

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const PortablePathSuffix = ".portable.json"

func LoadPortablesFromPaths(paths ...string) ([]GenericPortable, error) {
	ps := make([]GenericPortable, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(path, ".") {
			continue
		}
		if !strings.HasSuffix(path, PortablePathSuffix) {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading path: %s: %w", path, err)
		}
		p, err := LoadPortableFromJSON(b)
		if err != nil {
			return nil, fmt.Errorf("error loading portable from path: %s: %w", path, err)
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func LoadPortableFromJSON(b []byte) (GenericPortable, error) {
	var gp GenericPortable
	if err := json.Unmarshal(b, &gp); err != nil {
		return gp, err
	}
	return gp, nil
}

func NewPortableFactoriesFromItems(items ...GenericPortable) ([]*GenericPortableFactory, error) {
	m := map[int32][]*GenericPortable{}
	for _, item := range items {
		gps, err := NewGenericPortable(item)
		if err != nil {
			return nil, err
		}
		m[item.FID] = append(m[item.FID], gps)
	}
	fs := make([]*GenericPortableFactory, 0, len(items))
	for fid, ss := range m {
		f, err := NewGenericPortableFactory(fid, ss...)
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	return fs, nil
}
