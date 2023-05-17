package viridian

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"
)

type createClusterRequest struct {
	KubernetesClusterID int         `json:"kubernetesClusterId"`
	Name                string      `json:"name"`
	ClusterTypeID       ClusterType `json:"clusterTypeId"`
	PlanName            ClusterPlan `json:"planName"`
}

type createClusterResponse Cluster

func (a API) CreateCluster(ctx context.Context, name string, plan ClusterPlan, k8sClusterID int, isDev bool) (Cluster, error) {
	if name == "" {
		name = clusterName()
	}
	if plan == "" {
		plan = ClusterPlanServerless
	}
	ct := ClusterTypeProd
	if isDev {
		ct = ClusterTypeDev
	}
	c := createClusterRequest{
		KubernetesClusterID: k8sClusterID,
		Name:                name,
		ClusterTypeID:       ct,
		PlanName:            plan,
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
	cid, err := a.findClusterID(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/stop", cid), a.Token(), nil)
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
	cid, err := a.findClusterID(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/resume", cid), a.Token(), nil)
	if err != nil {
		return fmt.Errorf("resuming cluster: %w", err)
	}
	if !ok {
		return errors.New("could not resume the cluster")
	}
	return nil
}

func (a API) DeleteCluster(ctx context.Context, idOrName string) error {
	cid, err := a.findClusterID(ctx, idOrName)
	if err != nil {
		return err
	}
	err = doDelete(ctx, fmt.Sprintf("/cluster/%s", cid), a.Token())
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return nil
}

func (a API) GetCluster(ctx context.Context, idOrName string) (Cluster, error) {
	cid, err := a.findClusterID(ctx, idOrName)
	if err != nil {
		return Cluster{}, err
	}
	c, err := doGet[Cluster](ctx, fmt.Sprintf("/cluster/%s", cid), a.Token())
	if err != nil {
		return Cluster{}, fmt.Errorf("retrieving cluster: %w", err)
	}
	return c, nil
}

type ClusterTypeItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (a API) ListClusterTypes(ctx context.Context) ([]ClusterTypeItem, error) {
	csw, err := doGet[Wrapper[[]ClusterTypeItem]](ctx, "/cluster_types", a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing cluster types: %w", err)
	}
	return csw.Content, nil
}
