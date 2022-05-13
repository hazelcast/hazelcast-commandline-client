package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func FileExists(path string) (bool, error) {
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

func HZCHomePath() string {
	homeDirectoryPath, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("retrieving home directory: %w", err))
	}
	return filepath.Join(homeDirectoryPath, ".local/share/hz-cli")
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
