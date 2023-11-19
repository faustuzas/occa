package client

import (
	"context"
	"encoding/json"
	"fmt"

	pkghttp "github.com/faustuzas/tcha/src/pkg/http"
)

type Client struct {
	client *pkghttp.Client
	token  string
}

func New(address string) *Client {
	return &Client{client: pkghttp.NewClient(address)}
}

func (c *Client) CheckAvailability(ctx context.Context) error {
	resp, err := c.client.Get(ctx, "/health")
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if b := string(resp.Body); "ok" != b {
		return fmt.Errorf("gateway returned response %s", b)
	}

	return nil
}

func (c *Client) Authenticate(ctx context.Context, name string, password string) error {
	req := AuthenticationRequest{
		Username: name,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpResp, err := c.client.Post(ctx, "/authenticate", body)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if httpResp.StatusCode != 200 {
		return fmt.Errorf("server responded with status code %d and body: %s", httpResp.StatusCode, httpResp.Body)
	}

	var resp AuthenticationResponse
	if err = json.Unmarshal(httpResp.Body, &resp); err != nil {
		return fmt.Errorf("unmarshaling response: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("response from server: %v", resp.Error)
	}

	c.token = resp.Token

	return nil
}
