//go:build std || project

package project

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestCreateCommand(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		testHomeDir := "testdata/home"
		check.Must(paths.CopyDir(testHomeDir, tcx.HomePath()))
		tempDir := check.MustValue(os.MkdirTemp("", "clc-"))
		outDir := filepath.Join(tempDir, "my-project")
		fixtureDir := "testdata/fixture/simple"
		defer func() {
			// ignoring the error here
			_ = os.RemoveAll(outDir)
		}()
		ctx := context.Background()
		// logging to stderr in order to avoid creating the logs directory
		cmd := []string{"project", "create", "simple", "-o", outDir, "--log.path", "stderr", "another_key=foo", "key1=bar"}
		check.Must(tcx.CLC().Execute(ctx, cmd...))
		check.Must(compareDirectories(fixtureDir, outDir))
	})
}

func compareDirectories(dir1, dir2 string) error {
	hashes1, err := getDirectoryHashes(dir1)
	if err != nil {
		return err
	}
	hashes2, err := getDirectoryHashes(dir2)
	if err != nil {
		return err
	}
	return compareHashes(hashes1, hashes2)
}

func getDirectoryHashes(dir string) (map[string][16]byte, error) {
	hashes := make(map[string][16]byte)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileHash, err := calculateFileHash(path)
			if err != nil {
				return err
			}
			relativePath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			if filepath.Ext(relativePath) == keepExt || filepath.Ext(relativePath) == templateExt {
				relativePath, _ = paths.SplitExt(relativePath)
			}
			hashes[relativePath] = fileHash
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return hashes, nil
}

func calculateFileHash(file string) ([16]byte, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return [16]byte{}, err
	}
	return md5.Sum(b), nil
}

func compareHashes(hashes1, hashes2 map[string][16]byte) error {
	if len(hashes1) != len(hashes2) {
		return fmt.Errorf("number of files do not match")
	}
	for file, hash1 := range hashes1 {
		hash2, ok := hashes2[file]
		if !ok || hash1 != hash2 {
			return fmt.Errorf("hashes for %s are different", file)
		}
	}
	return nil
}
