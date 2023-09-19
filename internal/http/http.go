package http

import (
	"bytes"
	"context"
	"net/http"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{},
	}
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	buf := &bytes.Buffer{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, buf)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
