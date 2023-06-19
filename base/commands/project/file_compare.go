package project

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
			if filepath.Ext(relativePath) == keepExt || filepath.Ext(relativePath) == templateExt {
				relativePath = removeFileExt(relativePath)
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

func compareHashes(hashes1, hashes2 map[string]string) error {
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
