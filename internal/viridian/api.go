package viridian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvAPIBaseURL     = "HZ_CLOUD_COORDINATOR_BASE_URL"
	EnvAPIKey         = "CLC_VIRIDIAN_API_KEY"
	EnvAPISecret      = "CLC_VIRIDIAN_API_SECRET"
	EnvEmail          = "CLC_VIRIDIAN_EMAIL"
	EnvPassword       = "CLC_VIRIDIAN_PASSWORD"
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

func (a API) ListClusters(ctx context.Context) ([]Cluster, error) {
	csw, err := doGet[Wrapper[[]Cluster]](ctx, "/cluster", a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}
	return csw.Content, nil
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
	return res, fmt.Errorf("%d: %s", rawRes.StatusCode, string(rb))
}

func doPost[Req, Res any](ctx context.Context, path, token string, request Req) (res Res, err error) {
	m, err := json.Marshal(request)
	if err != nil {
		return res, fmt.Errorf("creating login payload: %w", err)
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
	rb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if res.StatusCode == 200 {
		return rb, nil
	}
	return nil, fmt.Errorf("%d: %s", res.StatusCode, string(rb))
}

func doCustomClassUpload(ctx context.Context, url, filePath, token string) error {
	reqBody := new(bytes.Buffer)
	w := multipart.NewWriter(reqBody)

	p, err := w.CreateFormFile("customClassesFile", filepath.Base(filePath))
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(p, file)
	if err != nil {
		return err
	}

	w.Close()

	req, err := http.NewRequest(http.MethodPost, makeUrl(url), reqBody)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	req = req.WithContext(ctx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	resBody := &bytes.Buffer{}
	_, err = resBody.ReadFrom(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%d: %s", res.StatusCode, resBody.String())
	}
	return nil
}

func doCustomClassDownload(ctx context.Context, url, className, token string) error {
	f, err := os.Create(className)
	if err != nil {
		return err
	}
	defer f.Close()

	req, err := http.NewRequest(http.MethodGet, makeUrl(url), nil)
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

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func APIClass() string {
	ac := os.Getenv(EnvAPI)
	if ac != "" {
		return ac
	}
	return "api"
}
