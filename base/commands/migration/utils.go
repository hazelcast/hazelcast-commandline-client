package migration

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
)

type bundleFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type configBundle struct {
	MigrationID    string       `json:"migrationId"`
	ConfigPath     string       `json:"configPath"`
	Source         []bundleFile `json:"source"`
	Target         []bundleFile `json:"target"`
	IMaps          []string     `json:"imaps"`
	ReplicatedMaps []string     `json:"replicatedMaps"`
}

func (cb *configBundle) Walk(root string) error {
	var err error
	cb.IMaps, err = readItems(filepath.Join(root, "data", "imap_names.txt"))
	if err != nil {
		return fmt.Errorf("reading IMaps: %w", err)
	}
	cb.ReplicatedMaps, err = readItems(filepath.Join(root, "data", "replicated_map_names.txt"))
	if err != nil {
		return fmt.Errorf("reading Replicated Maps: %w", err)
	}
	cb.Source, err = walkDir(filepath.Join(root, "source"))
	if err != nil {
		return fmt.Errorf("reading source directory")
	}
	cb.Target, err = walkDir(filepath.Join(root, "target"))
	if err != nil {
		return fmt.Errorf("reading target directory")
	}
	cb.ConfigPath, err = readPathAsString(filepath.Join(root, "data", "path.txt"))
	return nil
}

func walkDir(path string) ([]bundleFile, error) {
	if !paths.Exists(path) {
		return nil, nil
	}
	var bundle []bundleFile
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		b64 := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
		base64.StdEncoding.Encode(b64, b)
		bundle = append(bundle, bundleFile{
			Name:    d.Name(),
			Content: string(b64),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func readItems(path string) ([]string, error) {
	if !paths.Exists(path) {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// read items and put them in a set in order to make sure they are unique
	is := map[string]struct{}{}
	scn := bufio.NewScanner(f)
	for scn.Scan() {
		n := strings.TrimSpace(scn.Text())
		if n == "" {
			continue
		}
		is[n] = struct{}{}
	}
	items := mk.KeysOf(is)
	sort.Slice(items, func(i, j int) bool {
		return items[i] < items[j]
	})
	return items, nil
}

func readPathAsString(path string) (string, error) {
	if !paths.Exists(path) {
		return "", nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func MakeMigrationID() string {
	return types.NewUUID().String()
}
