//go:build std || viridian

package viridian

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const (
	stateRunning = "RUNNING"
	stateFailed  = "FAILED"
)

var (
	ErrClusterFailed  = errors.New("cluster failed")
	ErrLoadingSecrets = errors.New("could not load Viridian secrets, did you login?")
)

func findTokenPath(apiKey string) (string, error) {
	ac := viridian.APIClass()
	if apiKey == "" {
		apiKey = os.Getenv(viridian.EnvAPIKey)
	}
	if apiKey != "" {
		return fmt.Sprintf(viridian.FmtTokenFileName, ac, apiKey), nil
	}
	tokenPaths, err := findAll(secretPrefix)
	if err != nil {
		return "", fmt.Errorf("cannot access the secrets, did you login?: %w", err)
	}
	// sort tokens, so findTokenPath returns the same token everytime.
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
	if tp == "" {
		return "", fmt.Errorf("no secrets found, did you login?")
	}
	return tp, nil
}

func findAll(prefix string) ([]string, error) {
	return paths.FindAll(paths.Join(paths.Secrets(), prefix), func(basePath string, entry os.DirEntry) (ok bool) {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == filepath.Ext(viridian.FmtTokenFileName)
	})
}

func findKeyAndSecret(tokenPath string) (key, secret, apiBase string, err error) {
	key, _ = paths.SplitExt(tokenPath)
	key = strings.TrimPrefix(key, viridian.APIClass()+"-")
	fn := fmt.Sprintf(fmtSecretFileName, viridian.APIClass(), key)
	b, err := secrets.Read(secretPrefix, fn)
	if err != nil {
		return "", "", "", err
	}
	ss := string(b)
	// secret and API base
	ls := strings.SplitN(ss, "\n", 2)
	if len(ls) == 1 {
		return key, ls[0], "", nil
	}
	return key, ls[0], ls[1], nil
}

func getAPI(ec plug.ExecContext) (*viridian.API, error) {
	tp, err := findTokenPath(ec.Props().GetString(propAPIKey))
	if err != nil {
		return nil, err
	}
	ec.Logger().Info("Using Viridian secret at: %s", tp)
	token, err := secrets.Read(secretPrefix, tp)
	if err != nil {
		ec.Logger().Error(err)
		return nil, ErrLoadingSecrets
	}
	key, secret, base, err := findKeyAndSecret(tp)
	if err != nil {
		ec.Logger().Error(err)
		return nil, ErrLoadingSecrets
	}
	if base == "" {
		base = viridian.APIBaseURL()
	}
	return viridian.NewAPI(secretPrefix, key, secret, string(token), base), nil
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

func tryImportConfig(ctx context.Context, ec plug.ExecContext, api *viridian.API, clusterID, cfgName string) (configPath string, err error) {
	return importConfig(ctx, ec, api, clusterID, cfgName, "clc", config.CreateFromZip)
}

func importConfig(ctx context.Context, ec plug.ExecContext, api *viridian.API, clusterID, cfgName, language string, f func(context.Context, string, string, log.Logger) (string, error)) (configPath string, err error) {
	zipPath, stop, err := api.DownloadConfig(ctx, clusterID, language)
	if err != nil {
		return "", err
	}
	defer stop()
	return f(ctx, cfgName, zipPath, ec.Logger())
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

func ClusterType(isDev bool) string {
	if isDev {
		return "Development"
	}
	return "Production"
}
