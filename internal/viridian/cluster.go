package viridian

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

const ClusterTypeDevMode = "DEVMODE"
const ClusterTypeServerless = "SERVERLESS"

type createClusterRequest struct {
	KubernetesClusterID int    `json:"kubernetesClusterId"`
	Name                string `json:"name"`
	ClusterTypeID       int64  `json:"clusterTypeId"`
	PlanName            string `json:"planName"`
}

type createClusterResponse Cluster

func (a *API) CreateCluster(ctx context.Context, name string, clusterType string, k8sClusterID int, hzVersion string) (Cluster, error) {
	if name == "" {
		name = clusterName()
	}
	cType, err := a.FindClusterType(ctx, clusterType)
	if err != nil {
		return Cluster{}, err
	}
	clusterTypeID := cType.ID
	planName := cType.Name
	if strings.ToUpper(cType.Name) == ClusterTypeDevMode && hzVersion == "" {
		planName = ClusterTypeServerless
	}
	c := createClusterRequest{
		KubernetesClusterID: k8sClusterID,
		Name:                name,
		ClusterTypeID:       clusterTypeID,
		PlanName:            planName,
	}
	cluster, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Cluster, error) {
		u := a.makeURL("/cluster")
		c, err := doPost[createClusterRequest, createClusterResponse](ctx, u, a.Token, c)
		return Cluster(c), err
	})
	if err != nil {
		return Cluster{}, fmt.Errorf("creating cluster: %w", err)
	}
	return cluster, nil
}

func clusterName() string {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "cluster"
	}
	base := path.Base(cwd)
	date := time.Now().Format("2006-01-02-15-04-05")
	num := rand.Intn(9999)
	return fmt.Sprintf("%s-%s-%.4d", base, date, num)
}

func (a *API) StopCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (bool, error) {
		u := a.makeURL("/cluster/%s/stop", c.ID)
		return doPost[[]byte, bool](ctx, u, a.Token, nil)
	})
	if err != nil {
		return fmt.Errorf("stopping cluster: %w", err)
	}
	if !ok {
		return errors.New("could not stop the cluster")
	}
	return nil
}

func (a *API) ListClusters(ctx context.Context) ([]Cluster, error) {
	csw, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Wrapper[[]Cluster], error) {
		u := a.makeURL("/cluster?size=500")
		return doGet[Wrapper[[]Cluster]](ctx, u, a.Token)
	})
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}
	return csw.Content, nil
}

func (a *API) ResumeCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (bool, error) {
		u := a.makeURL("/cluster/%s/resume", c.ID)
		return doPost[[]byte, bool](ctx, u, a.Token, nil)
	})
	if err != nil {
		return fmt.Errorf("resuming cluster: %w", err)
	}
	if !ok {
		return errors.New("could not resume the cluster")
	}
	return nil
}

func (a *API) DeleteCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	_, err = RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (any, error) {
		u := a.makeURL("/cluster/%s", c.ID)
		err = doDelete(ctx, u, a.Token)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return nil
}

func (a *API) GetCluster(ctx context.Context, idOrName string) (Cluster, error) {
	cluster, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return Cluster{}, err
	}
	c, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Cluster, error) {
		u := a.makeURL("/cluster/%s", cluster.ID)
		return doGet[Cluster](ctx, u, a.Token)
	})
	if err != nil {
		return Cluster{}, fmt.Errorf("retrieving cluster: %w", err)
	}
	return c, nil
}

func (a *API) ListClusterTypes(ctx context.Context) ([]ClusterType, error) {
	csw, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Wrapper[[]ClusterType], error) {
		u := a.makeURL("/cluster_types")
		return doGet[Wrapper[[]ClusterType]](ctx, u, a.Token)
	})
	if err != nil {
		return nil, fmt.Errorf("listing cluster types: %w", err)
	}
	return csw.Content, nil
}

func (a *API) makeURL(format string, args ...any) string {
	var sb strings.Builder
	sb.WriteString(a.APIBaseURL)
	sb.WriteString(fmt.Sprintf(format, args...))
	return sb.String()
}
