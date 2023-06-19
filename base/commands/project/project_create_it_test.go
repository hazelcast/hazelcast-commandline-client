package project

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

const (
	templateName   = "my-template"
	outputDir      = "my-project"
	testProjectDir = "testdata/my-project"
)

func TestCreateCommand(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "project", "create", templateName, "--output-dir", outputDir, "myName=foo", "mySurname=bar"))
			check.MustValue(compareDirectories(outputDir, testProjectDir))
		})
	})
	teardown()
}

func compareDirectories(dir1, dir2 string) (bool, error) {
	hashes1, err := getDirectoryHashes(dir1)
	if err != nil {
		return false, err
	}

	hashes2, err := getDirectoryHashes(dir2)
	if err != nil {
		return false, err
	}

	return compareHashes(hashes1, hashes2), nil
}

func getDirectoryHashes(dir string) (map[string]string, error) {
	hashes := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileHash, err := getFileHash(path)
			if err != nil {
				return err
			}
			relativePath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
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

func getFileHash(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func compareHashes(hashes1, hashes2 map[string]string) bool {
	if len(hashes1) != len(hashes2) {
		return false
	}
	for file, hash1 := range hashes1 {
		hash2, ok := hashes2[file]
		if !ok || hash1 != hash2 {
			return false
		}
	}

	return true
}

func teardown() {
	os.RemoveAll(outputDir)
}
