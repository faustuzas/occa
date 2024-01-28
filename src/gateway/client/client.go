package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	pkghttp "github.com/faustuzas/occa/src/pkg/http"
)

const (
	heartBeatInterval = 10 * time.Second
)

type Client struct {
	client *pkghttp.Client

	token         string
	lastRequestAt atomic.Int64

	stopCh chan struct{}

	logger *zap.Logger
}

func New(address string, logger *zap.Logger) *Client {
	return &Client{
		client: pkghttp.NewClient(address),
		stopCh: make(chan struct{}),

		logger: logger,
	}
}

func (c *Client) CheckAvailability(ctx context.Context) error {
	resp, err := c.client.Get(ctx, "/health")
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if code := resp.StatusCode; code != http.StatusOK {
		return fmt.Errorf("gateway returned status code %d", code)
	}

	return nil
}

func (c *Client) Register(ctx context.Context, name string, password string) error {
	req := RegistrationRequest{
		Username: name,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpResp, err := c.client.Post(ctx, "/register", body)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if httpResp.StatusCode != 200 {
		return fmt.Errorf("server responded with status code %d and body: %s", httpResp.StatusCode, httpResp.Body)
	}

	var resp RegistrationResponse
	if err = json.Unmarshal(httpResp.Body, &resp); err != nil {
		return fmt.Errorf("unmarshaling response: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("response from server: %v", resp.Error)
	}

	return nil
}

func (c *Client) Login(ctx context.Context, name string, password string) error {
	req := LoginRequest{
		Username: name,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpResp, err := c.client.Post(ctx, "/login", body)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if httpResp.StatusCode != 200 {
		return fmt.Errorf("server responded with status code %d and body: %s", httpResp.StatusCode, httpResp.Body)
	}

	var resp LoginResponse
	if err = json.Unmarshal(httpResp.Body, &resp); err != nil {
		return fmt.Errorf("unmarshaling response: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("response from server: %v", resp.Error)
	}

	c.token = resp.Token
	c.markAsActive()

	go c.heartbeatLoop()

	return nil
}

func (c *Client) ActiveUsers(ctx context.Context) ([]string, error) {
	httpResp, err := c.client.GetWithHeaders(ctx, "/active-users", map[string]string{
		"Authorization": c.token,
	})
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("server responded with status code %d and body: %s", httpResp.StatusCode, httpResp.Body)
	}

	var resp ActiveUsersResponse
	if err = json.Unmarshal(httpResp.Body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	return resp.ActiveUsers, nil
}

// heartbeatLoop maintains the heartbeat with the gateway. Expects token set and not modified further on.
func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(heartBeatInterval)
	defer ticker.Stop()

	for {
		c.maybeHeartbeat()

		select {
		case <-ticker.C:
		case <-c.stopCh:
			return
		}
	}
}

func (c *Client) maybeHeartbeat() {
	if time.Now().Add(-heartBeatInterval).Unix() < c.lastRequestAt.Load() {
		return
	}

	if c.heartbeat() {
		c.markAsActive()
	}
}

func (c *Client) heartbeat() bool {
	resp, err := c.client.PostWithHeaders(context.Background(), "/heartbeat", nil, map[string]string{
		"Authorization": c.token,
	})
	if err != nil {
		c.logger.Warn("failed to heart beat to gateway", zap.Error(err))
		return false
	}

	if resp.StatusCode != 200 {
		c.logger.Warn("server responded with non-200 status code to the heartbeat",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(resp.Body)))
		return false
	}

	return true
}

func (c *Client) markAsActive() {
	c.lastRequestAt.Store(time.Now().Unix())
}

func (c *Client) Close() {
	if c.stopCh == nil {
		return
	}
	close(c.stopCh)
	c.stopCh = nil
}
