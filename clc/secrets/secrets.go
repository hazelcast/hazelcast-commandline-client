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
	tokenFileFormat        = "%s-%s"
	refreshTokenFileFormat = "%s-refresh-%s"
	expiresInFileFormat    = "%s-expiry-%s-%d"
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

func Save(secretPrefix, key, token, refreshToken string, expiresIn int) error {
	tokenFile := fmt.Sprintf(tokenFileFormat, viridian.APIClass(), key)
	refreshTokenFile := fmt.Sprintf(refreshTokenFileFormat, viridian.APIClass(), key)
	expiresInFile := fmt.Sprintf(fmt.Sprintf(expiresInFileFormat, viridian.APIClass(), key, time.Now().Unix()))
	if err := os.MkdirAll(paths.Secrets(), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	if err := Write(secretPrefix, tokenFile, []byte(token)); err != nil {
		return err
	}
	if err := Write(secretPrefix, refreshTokenFile, []byte(refreshToken)); err != nil {
		return err
	}
	oldExpiryFiles, err := paths.FindAll(paths.Join(paths.Secrets(), secretPrefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && strings.Contains(entry.Name(), "expiry") && strings.Contains(entry.Name(), key)
	})
	if err != nil {
		return err
	}
	for _, old := range oldExpiryFiles {
		err = os.Remove(paths.Join(paths.Secrets(), secretPrefix, old))
		if err != nil {
			return err
		}
	}
	if err = Write(secretPrefix, expiresInFile, []byte(strconv.Itoa(expiresIn))); err != nil {
		return err
	}
	return nil
}

func RefreshTokenIfExpired(secretPrefix, tokenFileName string) error {
	apiKey := strings.Split(tokenFileName, "-")[1]
	expiryFile, err := paths.FindAll(paths.Join(paths.Secrets(), secretPrefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && strings.Contains(entry.Name(), "expiry") && strings.Contains(entry.Name(), apiKey)
	})
	if err != nil {
		return err
	}
	expired, err := isTokenExpired(secretPrefix, apiKey)
	if err != nil {
		return err
	}
	if expired {
		refreshTokenFile, err := paths.FindAll(paths.Join(paths.Secrets(), secretPrefix), func(basePath string, entry os.DirEntry) (ok bool) {
			return !entry.IsDir() && strings.Contains(entry.Name(), "refresh") && strings.Contains(entry.Name(), apiKey)
		})
		if err != nil {
			return err
		}
		api := viridian.API{}
		refreshToken, err := Read(secretPrefix, refreshTokenFile[0])
		if err != nil {
			return err
		}
		refresh, err := api.RefreshAccessToken(context.Background(), string(refreshToken))
		if err != nil {
			return err
		}
		if err = Write(secretPrefix, tokenFileName, []byte(refresh.AccessToken)); err != nil {
			return err
		}
		if err = Write(secretPrefix, refreshTokenFile[0], []byte(refresh.RefreshToken)); err != nil {
			return err
		}
		// update expiry file's date
		expiresInFile := fmt.Sprintf(fmt.Sprintf(expiresInFileFormat, viridian.APIClass(), apiKey, time.Now().Unix()))
		err = os.Rename(paths.Join(paths.Secrets(), secretPrefix, expiryFile[0]), paths.Join(paths.Secrets(), secretPrefix, expiresInFile))
		if err != nil {
			return err
		}
	}
	return nil
}

func isTokenExpired(secretPrefix, expiryFileName string) (bool, error) {
	expiry, err := Read(secretPrefix, expiryFileName)
	if err != nil {
		return false, err
	}
	expiryDuration, err := strconv.Atoi(string(expiry))
	if err != nil {
		return false, err
	}
	ts := strings.Split(expiryFileName, "-")[len(strings.Split(expiryFileName, "-"))-1]
	i, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false, err
	}
	t := time.Unix(i, 0)
	if t.Add(time.Second * time.Duration(expiryDuration)).After(time.Now()) {
		return false, nil
	}
	return true, nil
}
