package portal

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	UserAgent      string
	client         *http.Client
	transport      *http.Transport
	OnHeaders      func(req *http.Request)
	AbortOnNone2xx bool
}

func GetDefaultTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
}

func NewDefaultClient(transport *http.Transport) *Client {
	if transport == nil {
		transport = GetDefaultTransport()
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	return &Client{
		client:         client,
		transport:      transport,
		UserAgent:      "gocanto.dev",
		OnHeaders:      nil,
		AbortOnNone2xx: false,
	}
}

func (f *Client) Get(ctx context.Context, url string) (string, error) {
	if f == nil || f.client == nil {
		return "", fmt.Errorf("client is nil")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if f.OnHeaders != nil {
		f.OnHeaders(req)
	}

	req.Header.Set("User-Agent", f.UserAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}

	defer resp.Body.Close()

	if f.AbortOnNone2xx && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return "", fmt.Errorf("received non-2xx status code: %d", resp.StatusCode)
	}

	body, err := ReadWithSizeLimit(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
