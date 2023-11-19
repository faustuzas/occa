package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

func NewClient(baseAddress string) *Client {
	return &Client{
		baseAddress: baseAddress,
		client:      http.DefaultClient,
	}
}

type Client struct {
	baseAddress string
	client      *http.Client
}

type Response struct {
	StatusCode int
	Body       []byte
}

func (c Client) Get(ctx context.Context, path string) (Response, error) {
	return c.execute(ctx, http.MethodGet, path, nil)
}

func (c Client) Post(ctx context.Context, path string, body []byte) (Response, error) {
	return c.execute(ctx, http.MethodPost, path, body)
}

func (c Client) execute(ctx context.Context, method, path string, body []byte) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.url(path), bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("creating request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("executing request: %w", err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, fmt.Errorf("reading response body: %w", err)
	}

	return Response{
		StatusCode: res.StatusCode,
		Body:       resBody,
	}, nil
}

func (c Client) url(path string) string {
	return fmt.Sprintf("http://%s%s", c.baseAddress, path)
}
