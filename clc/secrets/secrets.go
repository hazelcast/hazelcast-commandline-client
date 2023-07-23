package secrets

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

const (
	accessTokenFileFormat  = "%s-%s.access"
	refreshTokenFileFormat = "%s-%s.refresh"
	expiresInFileFormat    = "%s-%s.expiry"
)

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

func Save(secretPrefix, key, token, refreshToken string, expiresIn int) error {
	if err := os.MkdirAll(paths.Secrets(), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	if err := saveToken(secretPrefix, key, token); err != nil {
		return err
	}
	if err := saveRefreshToken(secretPrefix, key, refreshToken); err != nil {
		return err
	}
	if err := saveExpiry(secretPrefix, key, expiresIn); err != nil {
		return err
	}
	return nil
}
