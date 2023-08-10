package sse

import (
	"context"
	"errors"
	"net/http"
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

func (c *Client) Subscribe(ctx context.Context) (<-chan *Event, <-chan error) {
	errCh := make(chan error)
	evCh := make(chan *Event)
	go func() {
		defer close(errCh)
		defer close(evCh)
		resp, err := c.sendRequest(ctx)
		if err != nil {
			errCh <- ErrNoConnection
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			errCh <- ErrNoConnection
			return
		}
		reader := NewEventScanner(resp.Body)
		for {
			event, err := reader.ReadEvent()
			if err != nil {
				errCh <- err
				return
			}
			evCh <- event
		}
	}()
	return evCh, errCh
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
