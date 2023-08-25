package sse

import (
	"context"
	"errors"
	"net/http"

	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

var (
	ErrNoConnection = errors.New("no connection")
)

type Client struct {
	URL        string
	HTTPClient *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		URL:        url,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) SubscribeWithCallback(ctx context.Context, process func(*Event) error) error {
	resp, err := c.sendRequest(ctx)
	if err != nil {
		return ErrNoConnection
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return viridian.NewHTTPClientError(resp.StatusCode, nil)
	}
	reader := NewEventScanner(resp.Body)
	for {
		// can exit on context cancelation because uses response body
		event, err := reader.ReadEvent()
		if err != nil {
			return err
		}
		if err := process(event); err != nil {
			return err
		}
	}
}

func (c *Client) sendRequest(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")
	return c.HTTPClient.Do(req)
}
