package secrets

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

func Write(prefix, name, token string) (err error) {
	path := paths.ResolveSecretPath(prefix, name)
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening secrets file: %w", err)
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()
	b64 := base64.NewEncoder(base64.StdEncoding, f)
	// ignoring the error here
	defer b64.Close()
	if _, err = b64.Write([]byte(token)); err != nil {
		return fmt.Errorf("writing the secret: %w", err)
	}
	return nil
}

func Read(prefix, name string) (string, error) {
	f, err := os.Open(paths.ResolveSecretPath(prefix, name))
	if err != nil {
		return "", err
	}
	// ignoring the error here
	defer f.Close()
	b64 := base64.NewDecoder(base64.StdEncoding, f)
	b, err := io.ReadAll(b64)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func FindAll(prefix string) ([]string, error) {
	return paths.FindAll(paths.Join(paths.Secrets(), prefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir()
	})
}
