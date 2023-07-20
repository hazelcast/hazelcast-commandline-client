package viridian

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const (
	stateRunning = "RUNNING"
	stateFailed  = "FAILED"
)

var (
	ErrClusterFailed = errors.New("cluster failed")
)

func findToken(apiKey string) (string, error) {
	ac := viridian.APIClass()
	if apiKey == "" {
		apiKey = os.Getenv(viridian.EnvAPIKey)
	}
	if apiKey != "" {
		return fmt.Sprintf("%s-%s", ac, apiKey), nil
	}
	tokenPaths, err := secrets.FindAll(secretPrefix)
	if err != nil {
		return "", fmt.Errorf("cannot access the secrets, did you login?: %w", err)
	}
	// sort tokens, so findToken returns the same token everytime.
	sort.Slice(tokenPaths, func(i, j int) bool {
		return tokenPaths[i] < tokenPaths[j]
	})
	var tp string
	for _, p := range tokenPaths {
		if strings.HasPrefix(p, ac) {
			tp = p
			break
		}
	}
	token := strings.Split(tp, "-")[1]
	expired, err := isTokenExpired(token)
	if err != nil {
		return "", err
	}
	if expired {
		refreshTokenFile, err := paths.FindAll(paths.Join(paths.Secrets(), secretPrefix), func(basePath string, entry os.DirEntry) (ok bool) {
			return !entry.IsDir() && strings.Contains(entry.Name(), "refresh") && strings.Contains(entry.Name(), token)
		})
		if err != nil {
			return "", err
		}
		api := viridian.API{}
		refreshToken, err := secrets.Read(secretPrefix, refreshTokenFile[0])
		if err != nil {
			return "", err
		}
		refresh, err := api.RefreshAccessToken(context.Background(), string(refreshToken))
		if err != nil {
			return "", err
		}
		if err = secrets.Write(secretPrefix, tp, []byte(refresh.AccessToken)); err != nil {
			return "", err
		}
		if err = secrets.Write(secretPrefix, refreshTokenFile[0], []byte(refresh.RefreshToken)); err != nil {
			return "", err
		}
		// update expiry file's date
		expiryFile, err := paths.FindAll(paths.Join(paths.Secrets(), secretPrefix), func(basePath string, entry os.DirEntry) (ok bool) {
			return !entry.IsDir() && strings.Contains(entry.Name(), "expiry") && strings.Contains(entry.Name(), token)
		})
		if err != nil {
			return "", err
		}
		expiresInFile := fmt.Sprintf(fmt.Sprintf("%s-expiry-%s-%d", viridian.APIClass(), token, time.Now().Unix()))
		err = os.Rename(paths.Join(paths.Secrets(), secretPrefix, expiryFile[0]), paths.Join(paths.Secrets(), secretPrefix, expiresInFile))
		if err != nil {
			return "", err
		}
	}
	if tp == "" {
		return "", fmt.Errorf("no secrets found, did you login?")
	}
	return tp, nil
}

func isTokenExpired(token string) (bool, error) {
	expiryFile, err := paths.FindAll(paths.Join(paths.Secrets(), secretPrefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && strings.Contains(entry.Name(), "expiry") && strings.Contains(entry.Name(), token)
	})
	if err != nil {
		return false, err
	}
	expiry, err := secrets.Read(secretPrefix, expiryFile[0])
	if err != nil {
		return false, err
	}
	expiryDuration, err := strconv.Atoi(string(expiry))
	if err != nil {
		return false, err
	}
	ts := strings.Split(expiryFile[0], "-")[len(strings.Split(expiryFile[0], "-"))-1]
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

func getAPI(ec plug.ExecContext) (*viridian.API, error) {
	tp, err := findToken(ec.Props().GetString(propAPIKey))
	if err != nil {
		return nil, err
	}
	ec.Logger().Info("Using Viridian secret at: %s", tp)
	token, err := secrets.Read(secretPrefix, tp)
	if err != nil {
		ec.Logger().Error(err)
		return nil, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	return viridian.NewAPI(string(token)), nil
}

func waitClusterState(ctx context.Context, ec plug.ExecContext, api *viridian.API, clusterIDOrName, state string) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		cs, err := api.ListClusters(ctx)
		if err != nil {
			return err
		}
		for _, c := range cs {
			if c.ID != clusterIDOrName && c.Name != clusterIDOrName {
				continue
			}
			ok, err := matchClusterState(c, state)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			ec.Logger().Info("Waiting for cluster %s with state %s to transition to %s", clusterIDOrName, c.State, state)
			time.Sleep(2 * time.Second)
		}
	}
}

func matchClusterState(cluster viridian.Cluster, state string) (bool, error) {
	if cluster.State == state {
		return true, nil
	}
	if cluster.State == stateFailed {
		return false, ErrClusterFailed
	}
	return false, nil
}

func handleErrorResponse(ec plug.ExecContext, err error) error {
	ec.Logger().Error(err)
	// if it is a http client error, return the simplified error to user
	var ce viridian.HTTPClientError
	if errors.As(err, &ce) {
		if ce.Code() == http.StatusUnauthorized {
			return fmt.Errorf("authentication error, did you login?")
		}
		return fmt.Errorf(ce.Text())
	}
	// if it is not a http client error, return it directly as always
	return err
}

func fixClusterState(state string) string {
	// this is a temporary workaround until this is changed in the API
	state = strings.Replace(state, "STOPPED", "PAUSED", 1)
	state = strings.Replace(state, "STOP", "PAUSE", 1)
	return state
}
