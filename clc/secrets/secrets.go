package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const (
	tokenFileFormat        = "%s-%s.access"
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

func FindAllAccessTokens(prefix string) ([]string, error) {
	return paths.FindAll(paths.Join(paths.Secrets(), prefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() &&
			!strings.Contains(entry.Name(), "expiry") &&
			!strings.Contains(entry.Name(), "refresh")
	})
}

func FindRefreshToken(prefix, apiKey string) ([]string, error) {
	return paths.FindAll(paths.Join(paths.Secrets(), prefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && strings.Contains(entry.Name(), "refresh") && strings.Contains(entry.Name(), apiKey)
	})
}

func Save(secretPrefix, key, token, refreshToken string, expiresIn int) error {
	tokenFileName := fmt.Sprintf(tokenFileFormat, viridian.APIClass(), key)
	refreshTokenFileName := fmt.Sprintf(refreshTokenFileFormat, viridian.APIClass(), key)
	expiresInFileName := fmt.Sprintf(fmt.Sprintf(expiresInFileFormat, viridian.APIClass(), key))
	if err := os.MkdirAll(paths.Secrets(), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	if err := Write(secretPrefix, tokenFileName, []byte(token)); err != nil {
		return err
	}
	if err := Write(secretPrefix, refreshTokenFileName, []byte(refreshToken)); err != nil {
		return err
	}
	path := paths.ResolveSecretPath(secretPrefix, expiresInFileName)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	ex := strconv.Itoa(expiresIn)
	if err := os.WriteFile(path, []byte(strings.Join([]string{ts, ex}, "-")), 0600); err != nil {
		return fmt.Errorf("writing the expires in to file: %w", err)
	}
	return nil
}

func RefreshTokenIfExpired(secretPrefix, tokenFileName string) error {
	s := strings.TrimPrefix(tokenFileName, "api-")
	apiKey := strings.TrimSuffix(s, ".access")
	expiryFile := fmt.Sprintf("api-%s.expiry", apiKey)
	expired, err := isTokenExpired(secretPrefix, expiryFile)
	if err != nil {
		return err
	}
	if expired {
		refreshTokenFile := fmt.Sprintf("%s/%s/api-%s.refresh", paths.Secrets(), secretPrefix, apiKey)
		refreshToken, err := Read(secretPrefix, refreshTokenFile)
		if err != nil {
			return err
		}
		refresh, err := viridian.API{}.RefreshAccessToken(context.Background(), string(refreshToken))
		if err != nil {
			return err
		}
		if err = Write(secretPrefix, tokenFileName, []byte(refresh.AccessToken)); err != nil {
			return err
		}
		if err = Write(secretPrefix, refreshTokenFile, []byte(refresh.RefreshToken)); err != nil {
			return err
		}
		_, d, err := expiryValues(secretPrefix, expiryFile)
		if err != nil {
			return err
		}
		path := paths.ResolveSecretPath(secretPrefix, expiryFile)
		ts := strconv.FormatInt(time.Now().Unix(), 10)
		ex := strconv.Itoa(d)
		if err := os.WriteFile(path, []byte(strings.Join([]string{ts, ex}, "-")), 0600); err != nil {
			return fmt.Errorf("writing the expires in to file: %w", err)
		}
	}
	return nil
}

func isTokenExpired(secretPrefix, expiryFileName string) (bool, error) {
	t, d, err := expiryValues(secretPrefix, expiryFileName)
	if err != nil {
		return false, err
	}
	tt := time.Unix(t, 0)
	if tt.Add(time.Second * time.Duration(d)).After(time.Now()) {
		return false, nil
	}
	return true, nil
}

func expiryValues(secretPrefix, expiryFileName string) (int64, int, error) {
	expiry, err := os.ReadFile(paths.Join(paths.Secrets(), secretPrefix, expiryFileName))
	if err != nil {
		return 0, 0, err
	}
	ex := strings.Split(string(expiry), "-")
	expiryDuration, err := strconv.Atoi(string(ex[1]))
	if err != nil {
		return 0, 0, err
	}
	creationTime, err := strconv.ParseInt(ex[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return creationTime, expiryDuration, nil
}
