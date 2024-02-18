package etcd

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
)

type LeasedClient struct {
	client *clientv3.Client
	lease  clientv3.Lease

	mu       sync.Mutex
	leaseTTL time.Duration
	leaseID  clientv3.LeaseID
	leaseOK  bool

	leaseLostCh chan struct{}
	closeCh     chan struct{}

	i pkginstrument.Instrumentation
}

func NewLeasedClient(i pkginstrument.Instrumentation, client *clientv3.Client, leaseTTL time.Duration) *LeasedClient {
	return &LeasedClient{
		client: client,
		lease:  clientv3.NewLease(client),

		leaseTTL:    leaseTTL,
		leaseLostCh: make(chan struct{}),
		closeCh:     make(chan struct{}),

		i: i,
	}
}

// Start initializes a lease and tries to keep it alive indefinitely. To get a notification
// when the lease is lost, wait on channel returned by LeaseLostC.
func (c *LeasedClient) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.lease.Grant(ctx, int64(c.leaseTTL.Seconds()))
	if err != nil {
		return fmt.Errorf("granting lease: %w", err)
	}
	c.leaseID = resp.ID

	keepAliveCh, err := c.lease.KeepAlive(ctx, c.leaseID)
	if err != nil {
		return fmt.Errorf("keeping lease alive: %w", err)
	}

	c.leaseOK = true
	go func() {
		c.keepAlive(keepAliveCh)
	}()

	return nil
}

func (c *LeasedClient) LeaseLostC() <-chan struct{} {
	return c.leaseLostCh
}

func (c *LeasedClient) keepAlive(ch <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case <-c.closeCh:
		case _, ok := <-ch:
			if ok {
				continue
			}
		}

		if err := c.invalidateLease(context.Background()); err != nil {
			c.i.Logger.Error("failed revoking lease", zap.Error(err))
		}
		return
	}
}

func (c *LeasedClient) invalidateLease(ctx context.Context) error {
	c.mu.Lock()
	if !c.leaseOK {
		c.mu.Unlock()
		return fmt.Errorf("invalidating lease which is not valid")
	}

	c.leaseOK = false
	c.mu.Unlock()

	if _, err := c.lease.Revoke(ctx, c.leaseID); err != nil {
		return fmt.Errorf("revoking lease: %w", err)
	}
	close(c.leaseLostCh)

	return nil
}

func (c *LeasedClient) Close(ctx context.Context) error {
	close(c.closeCh)

	select {
	case <-c.leaseLostCh:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for the lease to be revoked")
	}
}

func (c *LeasedClient) Put(ctx context.Context, key string, value []byte) error {
	if _, err := c.client.Put(ctx, key, string(value), clientv3.WithLease(c.leaseID)); err != nil {
		return err
	}
	return nil
}
