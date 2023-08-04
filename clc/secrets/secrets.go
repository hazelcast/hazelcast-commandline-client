package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

const (
	TokenFileFormat  = "%s-%s.access"
	SecretFileFormat = "%s-%s.secret"
)

func Save(ctx context.Context, apiClass, secretPrefix, key, secret, token string) error {
	tokenFile := fmt.Sprintf(TokenFileFormat, apiClass, key)
	secretFile := fmt.Sprintf(SecretFileFormat, apiClass, key)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := os.MkdirAll(paths.Secrets(), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	if err := Write(secretPrefix, tokenFile, []byte(token)); err != nil {
		return err
	}
	if err := Write(secretPrefix, secretFile, []byte(secret)); err != nil {
		return err
	}
	return nil
}

func Write(prefix, name string, token []byte) (err error) {
	path := paths.ResolveSecretPath(prefix, name)
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(token)))
	base64.StdEncoding.Encode(b64, token)
	if err := os.WriteFile(path, b64, 0600); err != nil {
		return fmt.Errorf("writing the secret: %w", err)
	}
	return nil
}

func Read(prefix, name string) ([]byte, error) {
	b64, err := os.ReadFile(paths.ResolveSecretPath(prefix, name))
	if err != nil {
		return nil, fmt.Errorf("reading the secret: %w", err)
	}
	b := make([]byte, base64.StdEncoding.DecodedLen(len(b64)))
	n, err := base64.StdEncoding.Decode(b, b64)
	if err != nil {
		return nil, fmt.Errorf("decoding the secret: %w", err)
	}
	return b[:n], nil
}

func FindAll(prefix string) ([]string, error) {
	return paths.FindAll(paths.Join(paths.Secrets(), prefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == filepath.Ext(TokenFileFormat)
	})
}
