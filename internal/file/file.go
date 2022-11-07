package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

func Exists(path string) (bool, error) {
	var err error
	if _, err = os.Stat(path); err == nil {
		// conf file exists
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	// unexpected error
	return false, err
}

func HZCHomePath() (string, error) {
	if runtime.GOOS == "windows" {
		dir := os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
		// C:\Users\USERNAME\AppData\Local
		return filepath.Join(dir, "Local", "Hazelcast CLC"), nil
	}
	homeDirectoryPath, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("retrieving home directory: %w", err))
	}
	return filepath.Join(homeDirectoryPath, ".local/share/hz-cli"), nil
}

func CreateMissingDirsAndFileWithRWPerms(path string, content []byte) error {
	filePath, _ := filepath.Split(path)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return fmt.Errorf("can not create parent directories for file: %w", err)
	}
	if err := ioutil.WriteFile(path, content, 0600); err != nil {
		return fmt.Errorf("can not write to %s: %w", path, err)
	}
	return nil
}
