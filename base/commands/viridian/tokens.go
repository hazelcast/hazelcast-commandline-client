package viridian

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
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type Tokens struct {
	Key          string
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func FindTokens(ec plug.ExecContext) (Tokens, error) {
	apiKey := ec.Props().GetString(propAPIKey)
	accessTokenFilePath, err := findAccessTokenPath(apiKey)
	if err != nil {
		return Tokens{}, err
	}
	if apiKey == "" {
		apiKey = findAPIKey(accessTokenFilePath)
	}
	ec.Logger().Info("Using Viridian secret at: %s", accessTokenFilePath)
	err = refreshTokenIfExpired(secretPrefix, accessTokenFilePath)
	if err != nil {
		return Tokens{}, err
	}
	token, err := readToken(secretPrefix, apiKey, secrets.AccessTokenFileFormat)
	if err != nil {
		ec.Logger().Error(err)
		return Tokens{}, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	refToken, err := readToken(secretPrefix, apiKey, secrets.RefreshTokenFileFormat)
	if err != nil {
		ec.Logger().Error(err)
		return Tokens{}, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	_, expiresIn, err := expiryValues(secretPrefix, fmt.Sprintf(secrets.ExpiresInFileFormat, viridian.APIClass(), apiKey))
	if err != nil {
		ec.Logger().Error(err)
		return Tokens{}, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	return Tokens{
		Key:          apiKey,
		AccessToken:  token,
		RefreshToken: refToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func findAccessTokenPath(apiKey string) (string, error) {
	var path string
	if apiKey != "" {
		path = fmt.Sprintf(secrets.AccessTokenFileFormat, viridian.APIClass(), apiKey)
	} else if os.Getenv(viridian.EnvAPIKey) != "" {
		path = fmt.Sprintf(secrets.AccessTokenFileFormat, viridian.APIClass(), os.Getenv(viridian.EnvAPIKey))
	} else {
		tokenPaths, err := findAllAccessTokenFiles(secretPrefix)
		if err != nil {
			return "", fmt.Errorf("cannot access the secrets, did you login?: %w", err)
		}
		// sort tokens, so it returns the same token everytime.
		sort.Slice(tokenPaths, func(i, j int) bool {
			return tokenPaths[i] < tokenPaths[j]
		})
		for _, p := range tokenPaths {
			if strings.HasPrefix(p, viridian.APIClass()) {
				path = p
				break
			}
		}
		if path == "" {
			return "", fmt.Errorf("no secrets found, did you login?")
		}
	}
	return path, nil
}

func findAllAccessTokenFiles(prefix string) ([]string, error) {
	return paths.FindAll(paths.Join(paths.Secrets(), prefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == filepath.Ext(secrets.AccessTokenFileFormat)
	})
}

func readToken(prefix, apiKey, fileFormat string) (string, error) {
	b, err := secrets.Read(prefix, fmt.Sprintf(fileFormat, viridian.APIClass(), apiKey))
	return string(b), err
}

func refreshTokenIfExpired(secretPrefix, accessTokenFileName string) error {
	apiKey := findAPIKey(accessTokenFileName)
	expiryFileName := fmt.Sprintf(secrets.ExpiresInFileFormat, viridian.APIClass(), apiKey)
	expired, err := isTokenExpired(secretPrefix, expiryFileName)
	if err != nil {
		return err
	}
	if expired {
		refreshTokenFileName := fmt.Sprintf(secrets.RefreshTokenFileFormat, viridian.APIClass(), apiKey)
		refreshToken, err := secrets.Read(secretPrefix, refreshTokenFileName)
		if err != nil {
			return err
		}
		_, e, err := expiryValues(secretPrefix, expiryFileName)
		if err != nil {
			return err
		}
		api := viridian.API{Key: apiKey, SecretPrefix: secretPrefix, RefreshToken: string(refreshToken), ExpiresIn: e}
		r, err := api.RefreshAccessToken(context.Background())
		if err != nil {
			return err
		}
		// save to .access file
		if err = secrets.Write(secretPrefix, accessTokenFileName, []byte(r.AccessToken)); err != nil {
			return err
		}
		// save to .refresh file
		if err = secrets.Write(secretPrefix, refreshTokenFileName, []byte(r.RefreshToken)); err != nil {
			return err
		}
		// save to .expiry file
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
	pre := fmt.Sprintf("%s-", viridian.APIClass())
	ext := fmt.Sprintf("%s", filepath.Ext(secrets.AccessTokenFileFormat))
	return strings.TrimPrefix(strings.TrimSuffix(accessTokenFileName, ext), pre)
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
	expiryTime, err := strconv.ParseInt(ex[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return expiryTime, expiryDuration, nil
}

func expiryData(expiresIn int) ([]byte, error) {
	ts := strconv.FormatInt(secrets.CalculateExpiry(expiresIn), 10)
	ex := strconv.Itoa(expiresIn)
	return []byte(fmt.Sprintf("%s-%s", ts, ex)), nil
}
