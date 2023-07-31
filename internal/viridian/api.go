package viridian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	EnvAPIBaseURL     = "HZ_CLOUD_COORDINATOR_BASE_URL"
	EnvAPIKey         = "CLC_VIRIDIAN_API_KEY"
	EnvAPISecret      = "CLC_VIRIDIAN_API_SECRET"
	EnvAPI            = "CLC_EXPERIMENTAL_VIRIDIAN_API"
	DefaultAPIBaseURL = "https://api.viridian.hazelcast.com"
)

type Wrapper[T any] struct {
	Content T
}

type API struct {
	token string
}

func NewAPI(token string) *API {
	return &API{token: token}
}

func (a API) Token() string {
	return a.token
}

func (a API) ListAvailableK8sClusters(ctx context.Context) ([]K8sCluster, error) {
	c, err := doGet[[]K8sCluster](ctx, "/kubernetes_clusters/available", a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing available Kubernetes clusters: %w", err)
	}
	return c, nil
}

func (a API) ListCustomClasses(ctx context.Context, cluster string) ([]CustomClass, error) {
	c, err := a.FindCluster(ctx, cluster)
	if err != nil {
		return nil, err
	}
	csw, err := doGet[[]CustomClass](ctx, fmt.Sprintf("/cluster/%s/custom_classes", c.ID), a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing custom classes: %w", err)
	}
	return csw, nil
}

func (a API) UploadCustomClasses(ctx context.Context, p func(progress float32), cluster, filePath string) error {
	c, err := a.FindCluster(ctx, cluster)
	if err != nil {
		return err
	}
	err = doCustomClassUpload(ctx, p, fmt.Sprintf("/cluster/%s/custom_classes", c.ID), filePath, a.Token())
	if err != nil {
		return fmt.Errorf("uploading custom class: %w", err)
	}
	return nil
}

func (a API) DownloadCustomClass(ctx context.Context, p func(progress float32), targetInfo TargetInfo, cluster, artifact string) error {
	c, err := a.FindCluster(ctx, cluster)
	if err != nil {
		return err
	}
	artifactID, artifactName, err := a.findArtifactIDAndName(ctx, cluster, artifact)
	if err != nil {
		return err
	}
	if artifactID == 0 {
		return fmt.Errorf("no custom class artifact found with name or ID %s in cluster %s", artifact, c.ID)
	}
	url := fmt.Sprintf("/cluster/%s/custom_classes/%d", c.ID, artifactID)
	err = doCustomClassDownload(ctx, p, targetInfo, url, artifactName, a.token)
	if err != nil {
		return err
	}
	return nil
}

func (a API) DeleteCustomClass(ctx context.Context, cluster string, artifact string) error {
	c, err := a.FindCluster(ctx, cluster)
	if err != nil {
		return err
	}
	artifactID, _, err := a.findArtifactIDAndName(ctx, cluster, artifact)
	if err != nil {
		return err
	}
	if artifactID == 0 {
		return fmt.Errorf("no custom class artifact found with name or ID %s in cluster %s", artifact, c.ID)
	}
	err = doDelete(ctx, fmt.Sprintf("/cluster/%s/custom_classes/%d", c.ID, artifactID), a.token)
	if err != nil {
		return err
	}
	return nil
}

func (a API) DownloadConfig(ctx context.Context, clusterID string) (path string, stop func(), err error) {
	url := makeConfigURL(clusterID)
	return download(ctx, url, a.token)
}

func (a API) FindCluster(ctx context.Context, idOrName string) (Cluster, error) {
	clusters, err := a.ListClusters(ctx)
	if err != nil {
		return Cluster{}, err
	}
	for _, c := range clusters {
		if c.ID == idOrName || c.Name == idOrName {
			return c, nil
		}
	}
	return Cluster{}, fmt.Errorf("no such cluster found: %s", idOrName)
}

func (a API) FindClusterType(ctx context.Context, name string) (ClusterType, error) {
	cts, err := a.ListClusterTypes(ctx)
	if err != nil {
		return ClusterType{}, err
	}
	for _, ct := range cts {
		if strings.ToUpper(ct.Name) == strings.ToUpper(name) {
			return ct, nil
		}
	}
	return ClusterType{}, fmt.Errorf("no such cluster type found: %s", name)
}

func (a API) StreamLogs(ctx context.Context, idOrName string, out io.Writer) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/cluster/%s/logstream", c.ID)
	r, err := doGetRaw(ctx, path, a.token)
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(out, r)
	return err
}

func (a API) findArtifactIDAndName(ctx context.Context, clusterName, artifact string) (int64, string, error) {
	customClasses, err := a.ListCustomClasses(ctx, clusterName)
	if err != nil {
		return 0, "", err
	}
	var artifactName string
	var artifactID int64
	for _, cc := range customClasses {
		if cc.Name == artifact || strconv.FormatInt(cc.ID, 10) == artifact {
			artifactName = cc.Name
			artifactID = cc.ID
			break
		}
	}
	return artifactID, artifactName, nil
}

func APIBaseURL() string {
	u := os.Getenv(EnvAPIBaseURL)
	if u != "" {
		return u
	}
	return DefaultAPIBaseURL
}

func makeUrl(path string) string {
	path = strings.TrimLeft(path, "/")
	path = "/" + path
	return APIBaseURL() + path
}

func doGet[Res any](ctx context.Context, path, token string) (res Res, err error) {
	req, err := http.NewRequest(http.MethodGet, makeUrl(path), nil)
	if err != nil {
		return res, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	rawRes, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, fmt.Errorf("sending request: %w", err)
	}
	defer rawRes.Body.Close()
	rb, err := io.ReadAll(rawRes.Body)
	if err != nil {
		return res, fmt.Errorf("reading response: %w", err)
	}
	if rawRes.StatusCode == 200 {
		if err = json.Unmarshal(rb, &res); err != nil {
			return res, err
		}
		return res, nil
	}
	return res, NewHTTPClientError(rawRes.StatusCode, rb)
}

func doGetRaw(ctx context.Context, path, token string) (res io.ReadCloser, err error) {
	req, err := http.NewRequest(http.MethodGet, makeUrl(path), nil)
	if err != nil {
		return res, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	rawRes, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, fmt.Errorf("sending request: %w", err)
	}
	if rawRes.StatusCode == 200 {
		return rawRes.Body, nil
	}
	defer rawRes.Body.Close()
	rb, err := io.ReadAll(rawRes.Body)
	if err != nil {
		return res, fmt.Errorf("reading response: %w", err)
	}
	return res, NewHTTPClientError(rawRes.StatusCode, rb)
}

func doPost[Req, Res any](ctx context.Context, path, token string, request Req) (res Res, err error) {
	m, err := json.Marshal(request)
	if err != nil {
		return res, fmt.Errorf("creating payload: %w", err)
	}
	b, err := doPostBytes(ctx, makeUrl(path), token, m)
	if err != nil {
		return res, err
	}
	if err = json.Unmarshal(b, &res); err != nil {
		return res, err
	}
	return res, nil
}

func doPostBytes(ctx context.Context, url, token string, body []byte) ([]byte, error) {
	reader := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()
	rb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if res.StatusCode == 200 {
		return rb, nil
	}
	return nil, NewHTTPClientError(res.StatusCode, rb)
}

func doDelete(ctx context.Context, path, token string) error {
	req, err := http.NewRequest(http.MethodDelete, makeUrl(path), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()
	rb, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return NewHTTPClientError(res.StatusCode, rb)
	}
	return nil
}

func download(ctx context.Context, url, token string) (downloadPath string, stop func(), err error) {
	f, err := os.CreateTemp("", "clc-download-*")
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", nil, fmt.Errorf("creating request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	rawRes, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("sending request: %w", err)
	}
	defer rawRes.Body.Close()
	if rawRes.StatusCode == 200 {
		if _, err := io.Copy(f, rawRes.Body); err != nil {
			return "", nil, fmt.Errorf("downloading file: %w", err)
		}
		stop = func() {
			// ignoring tne error
			os.Remove(f.Name())
		}
		return f.Name(), stop, nil
	}
	rb, err := io.ReadAll(rawRes.Body)
	if err != nil {
		return "", nil, fmt.Errorf("reading error response: %w", err)
	}
	return "", nil, fmt.Errorf("%d: %s", rawRes.StatusCode, string(rb))
}

func makeConfigURL(clusterID string) string {
	return makeUrl(fmt.Sprintf("/client_samples/%s/python?source_identifier=default", clusterID))
}

func APIClass() string {
	ac := os.Getenv(EnvAPI)
	if ac != "" {
		return ac
	}
	return "api"
}
