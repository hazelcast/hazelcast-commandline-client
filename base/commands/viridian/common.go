package viridian

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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

func getAPI(ec plug.ExecContext) (*viridian.API, error) {
	t, err := FindTokens(ec)
	if err != nil {
		return nil, err
	}
	return viridian.NewAPI(secretPrefix, t.Key, t.Token, t.RefreshToken, t.ExpiresIn), nil
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
