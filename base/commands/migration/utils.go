package migration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func bundleDirAsJSON(root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	prefix := root
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	prefixLen := len(prefix)
	bundle := map[string]string{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
		bundle[path[prefixLen:]] = string(b64)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("bundling %s: %w", root, err)
	}
	b, err := json.Marshal(bundle)
	if err != nil {
		return "", fmt.Errorf("creating JSON document %s: %w", root, err)
	}
	return string(b), nil
}
