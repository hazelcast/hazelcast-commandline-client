package secrets

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type Tokens struct {
	Token        string
	RefreshToken string
	ExpiresIn    int
}

func FindTokens(ec plug.ExecContext, secretPrefix, propAPIKey string) (Tokens, error) {
	accessTokenFilePath, err := findAccessTokenFile(ec, secretPrefix, propAPIKey)
	if err != nil {
		return Tokens{}, err
	}
	apiKey := findAPIKey(accessTokenFilePath)
	ec.Logger().Info("Using Viridian secret at: %s", accessTokenFilePath)
	err = refreshTokenIfExpired(secretPrefix, accessTokenFilePath)
	if err != nil {
		return Tokens{}, err
	}
	token, err := readToken(secretPrefix, apiKey, accessTokenFileFormat)
	if err != nil {
		ec.Logger().Error(err)
		return Tokens{}, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	refToken, err := readToken(secretPrefix, apiKey, refreshTokenFileFormat)
	if err != nil {
		ec.Logger().Error(err)
		return Tokens{}, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	_, expiresIn, err := expiryValues(secretPrefix, fmt.Sprintf(expiresInFileFormat, viridian.APIClass(), apiKey))
	if err != nil {
		ec.Logger().Error(err)
		return Tokens{}, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	return Tokens{
		Token:        token,
		RefreshToken: refToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func findAccessTokenFile(ec plug.ExecContext, secretPrefix, propAPIKey string) (string, error) {
	tokenPaths, err := findAllAccessTokenFiles(secretPrefix)
	if err != nil {
		return "", fmt.Errorf("cannot access the secrets, did you login?: %w", err)
	}
	var accessTokenFilePath string
	if ec.Props().GetString(propAPIKey) != "" {
		accessTokenFilePath = fmt.Sprintf(accessTokenFileFormat, viridian.APIClass(), ec.Props().GetString(propAPIKey))
	} else if os.Getenv(viridian.EnvAPIKey) != "" {
		accessTokenFilePath = fmt.Sprintf(accessTokenFileFormat, viridian.APIClass(), os.Getenv(viridian.EnvAPIKey))
	} else {
		// sort tokens, so it returns the same token everytime.
		sort.Slice(tokenPaths, func(i, j int) bool {
			return tokenPaths[i] < tokenPaths[j]
		})
		for _, p := range tokenPaths {
			if strings.HasPrefix(p, viridian.APIClass()) {
				accessTokenFilePath = p
				break
			}
		}
		if accessTokenFilePath == "" {
			return "", fmt.Errorf("no secrets found, did you login?")
		}
	}
	return accessTokenFilePath, nil
}

func findAllAccessTokenFiles(prefix string) ([]string, error) {
	return paths.FindAll(paths.Join(paths.Secrets(), prefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == ".access"
	})
}

func readToken(prefix, apiKey, fileFormat string) (string, error) {
	b, err := Read(prefix, fmt.Sprintf(fileFormat, viridian.APIClass(), apiKey))
	return string(b), err
}

func refreshTokenIfExpired(secretPrefix, accessTokenFileName string) error {
	apiKey := findAPIKey(accessTokenFileName)
	expiryFileName := fmt.Sprintf(expiresInFileFormat, viridian.APIClass(), apiKey)
	expired, err := isTokenExpired(secretPrefix, expiryFileName)
	if err != nil {
		return err
	}
	if expired {
		refreshTokenFileName := fmt.Sprintf(refreshTokenFileFormat, viridian.APIClass(), apiKey)
		refreshToken, err := Read(secretPrefix, refreshTokenFileName)
		if err != nil {
			return err
		}
		r, err := viridian.API{RefreshToken: string(refreshToken)}.RefreshAccessToken(context.Background())
		if err != nil {
			return err
		}
		// save to .access file
		if err = Write(secretPrefix, accessTokenFileName, []byte(r.AccessToken)); err != nil {
			return err
		}
		// save to .refresh file
		if err = Write(secretPrefix, refreshTokenFileName, []byte(r.RefreshToken)); err != nil {
			return err
		}
		// save to .expiry file
		_, e, err := expiryValues(secretPrefix, expiryFileName)
		if err != nil {
			return err
		}
		d, err := expiryData(e)
		if err != nil {
			return err
		}
		if err := os.WriteFile(paths.ResolveSecretPath(secretPrefix, expiryFileName), d, 0600); err != nil {
			return fmt.Errorf("writing the expires in to file: %w", err)
		}
	}
	return nil
}

func findAPIKey(accessTokenFileName string) string {
	return strings.TrimPrefix(strings.TrimSuffix(accessTokenFileName, ".access"), "api-")
}

func isTokenExpired(secretPrefix, expiryFileName string) (bool, error) {
	t, _, err := expiryValues(secretPrefix, expiryFileName)
	if err != nil {
		return false, err
	}
	if time.Unix(t, 0).After(time.Now()) {
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
	expiryDuration, err := strconv.Atoi(ex[1])
	if err != nil {
		return 0, 0, err
	}
	creationTime, err := strconv.ParseInt(ex[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return creationTime, expiryDuration, nil
}

func expiryData(expiresIn int) ([]byte, error) {
	calcExpiry(expiresIn)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	ex := strconv.Itoa(expiresIn)
	return []byte(fmt.Sprintf("%s-%s", ts, ex)), nil
}

func calcExpiry(expiresIn int) int64 {
	return time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
}

func saveToken(secretPrefix, key, token string) error {
	fn := fmt.Sprintf(accessTokenFileFormat, viridian.APIClass(), key)
	if err := Write(secretPrefix, fn, []byte(token)); err != nil {
		return err
	}
	return nil
}

func saveRefreshToken(secretPrefix, key, refreshToken string) error {
	fn := fmt.Sprintf(refreshTokenFileFormat, viridian.APIClass(), key)
	if err := Write(secretPrefix, fn, []byte(refreshToken)); err != nil {
		return err
	}
	return nil
}

func saveExpiry(secretPrefix, key string, expiresIn int) error {
	fn := fmt.Sprintf(fmt.Sprintf(expiresInFileFormat, viridian.APIClass(), key))
	path := paths.ResolveSecretPath(secretPrefix, fn)
	ts := strconv.FormatInt(calcExpiry(expiresIn), 10)
	ex := strconv.Itoa(expiresIn)
	// We have to save to this file in (expireTime + expireDuration)-expireDuration format,
	// Because Viridian refresh token endpoint does not return expiryDuration
	// On Viridian expiryDuration is related to api key
	if err := os.WriteFile(path, []byte(fmt.Sprintf("%s-%s", ts, ex)), 0600); err != nil {
		return fmt.Errorf("writing the expires in to file: %w", err)
	}
	return nil
}
