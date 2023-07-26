package project

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestCreateCommand(t *testing.T) {
	os.Setenv(envTemplateSource, "https://github.com/kutluhanmetin")
	testCases := []struct {
		inputTemplateName string
		inputOutputDir    string
		inputArgs         []string
		testProjectDir    string
	}{
		{
			inputTemplateName: "simple-streaming-pipeline-template",
			inputOutputDir:    "my-simple-streaming-pipeline",
			inputArgs:         []string{"rootProjectName=simple-streaming-pipeline"},
			testProjectDir:    "../../../examples/platform/simple-streaming-pipeline",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.inputTemplateName, func(t *testing.T) {
			defer teardown(tc.inputOutputDir)
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				ctx := context.Background()
				tcx.WithReset(func() {
					cmd := []string{"project", "create", tc.inputTemplateName, tc.inputOutputDir}
					cmd = append(cmd, tc.inputArgs...)
					check.Must(tcx.CLC().Execute(ctx, cmd...))
				})
				tcx.WithReset(func() {
					check.Must(compareDirectories(tc.inputOutputDir, tc.testProjectDir))
				})
			})
		})
	}
}

func teardown(dir string) {
	os.RemoveAll(dir)
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
				relativePath = paths.SplitExt(relativePath)
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
	fmt.Println(fmt.Sprintf("%s file content:", file))
	fmt.Println(string(b))
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
