package viridian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	EnvAPIBaseURL     = "CLC_EXPERIMENTAL_VIRIDIAN_BASE_URL"
	EnvAPIKey         = "CLC_VIRIDIAN_API_KEY"
	EnvAPISecret      = "CLC_VIRIDIAN_API_SECRET"
	DefaultAPIBaseURL = "https://api.viridian.hazelcast.com"
)

type Wrapper[T any] struct {
	Content T
}

type API struct {
	token string
}

func NewAPI(token string) API {
	return API{token: token}
}

func (a API) Token() string {
	return a.token
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

func doGet(ctx context.Context, url, token string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
