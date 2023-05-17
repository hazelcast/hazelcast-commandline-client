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

type createClusterRequest struct {
	KubernetesClusterID int    `json:"kubernetesClusterId"`
	Name                string `json:"name"`
	ClusterTypeID       int64  `json:"clusterTypeId"`
	PlanName            string `json:"planName"`
}

type createClusterResponse Cluster

const defaultClusterType = "devmode"

func (a API) CreateCluster(ctx context.Context, name string, clusterType string, k8sClusterID int) (Cluster, error) {
	if name == "" {
		name = clusterName()
	}
	cType, err := a.FindClusterType(ctx, clusterType)
	if err != nil {
		return Cluster{}, err
	}
	clusterTypeID := cType.ID
	planName := cType.Name
	if cType.Name == "" {
		clusterTypes, err := a.ListClusterTypes(ctx)
		if err != nil {
			return Cluster{}, err
		}
		for _, ct := range clusterTypes {
			if strings.ToLower(ct.Name) == defaultClusterType {
				clusterTypeID = ct.ID
				planName = ct.Name
				break
			}
		}
	}
	c := createClusterRequest{
		KubernetesClusterID: k8sClusterID,
		Name:                name,
		ClusterTypeID:       clusterTypeID,
		PlanName:            planName,
	}
	cluster, err := doPost[createClusterRequest, createClusterResponse](ctx, "/cluster", a.Token(), c)
	if err != nil {
		return Cluster{}, fmt.Errorf("creating cluster: %w", err)
	}
	return Cluster(cluster), nil
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

func (a API) StopCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/stop", c.ID), a.Token(), nil)
	if err != nil {
		return fmt.Errorf("stopping cluster: %w", err)
	}
	if !ok {
		return errors.New("could not stop the cluster")
	}
	return nil
}

func (a API) ListClusters(ctx context.Context) ([]Cluster, error) {
	csw, err := doGet[Wrapper[[]Cluster]](ctx, "/cluster", a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}
	return csw.Content, nil
}

func (a API) ResumeCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/resume", c.ID), a.Token(), nil)
	if err != nil {
		return fmt.Errorf("resuming cluster: %w", err)
	}
	if !ok {
		return errors.New("could not resume the cluster")
	}
	return nil
}

func (a API) DeleteCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	err = doDelete(ctx, fmt.Sprintf("/cluster/%s", c.ID), a.Token())
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return nil
}

func (a API) GetCluster(ctx context.Context, idOrName string) (Cluster, error) {
	cluster, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return Cluster{}, err
	}
	c, err := doGet[Cluster](ctx, fmt.Sprintf("/cluster/%s", cluster.ID), a.Token())
	if err != nil {
		return Cluster{}, fmt.Errorf("retrieving cluster: %w", err)
	}
	return c, nil
}

type ClusterType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (a API) ListClusterTypes(ctx context.Context) ([]ClusterType, error) {
	csw, err := doGet[Wrapper[[]ClusterType]](ctx, "/cluster_types", a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing cluster types: %w", err)
	}
	return csw.Content, nil
}
