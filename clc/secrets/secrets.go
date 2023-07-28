package secrets

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

const (
	AccessTokenFileFormat  = "%s-%s.access"
	RefreshTokenFileFormat = "%s-%s.refresh"
	ExpiresInFileFormat    = "%s-%s.expiry"
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

func Save(secretPrefix, key, apiClass, token, refreshToken string, expiresIn int) error {
	if err := os.MkdirAll(paths.Secrets(), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	if err := saveToken(secretPrefix, apiClass, key, token); err != nil {
		return err
	}
	if err := saveRefreshToken(secretPrefix, apiClass, key, refreshToken); err != nil {
		return err
	}
	if err := saveExpiry(secretPrefix, apiClass, key, expiresIn); err != nil {
		return err
	}
	return nil
}

func saveToken(secretPrefix, apiClass, key, token string) error {
	fn := fmt.Sprintf(AccessTokenFileFormat, apiClass, key)
	if err := Write(secretPrefix, fn, []byte(token)); err != nil {
		return err
	}
	return nil
}

func saveRefreshToken(secretPrefix, apiClass, key, refreshToken string) error {
	fn := fmt.Sprintf(RefreshTokenFileFormat, apiClass, key)
	if err := Write(secretPrefix, fn, []byte(refreshToken)); err != nil {
		return err
	}
	return nil
}

func saveExpiry(secretPrefix, apiClass, key string, expiresIn int) error {
	fn := fmt.Sprintf(fmt.Sprintf(ExpiresInFileFormat, apiClass, key))
	path := paths.ResolveSecretPath(secretPrefix, fn)
	d, err := makeExpiryData(expiresIn)
	if err != nil {
		return err
	}
	// We have to save to this file in (expireTime + expireDuration)-expireDuration format,
	// Because Viridian refresh token endpoint does not return expiryDuration
	// On Viridian expiryDuration is related to api key
	if err := os.WriteFile(path, d, 0600); err != nil {
		return fmt.Errorf("writing the expires in to file: %w", err)
	}
	return nil
}

func makeExpiryData(expiresIn int) ([]byte, error) {
	ts := strconv.FormatInt(calculateExpiry(expiresIn), 10)
	ex := strconv.Itoa(expiresIn)
	return []byte(fmt.Sprintf("%s-%s", ts, ex)), nil
}

func calculateExpiry(expiresIn int) int64 {
	return time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
}
