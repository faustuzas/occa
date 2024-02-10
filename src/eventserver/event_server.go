package eventserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

type EventServer interface {
	SendEvent(ctx context.Context, msg Event) error
	ServeConnection(id pkgid.ID, connection Connection) error
}

func NewEventServer(l *zap.Logger, r prometheus.Registerer) (EventServer, error) {
	return &eventServer{
		logger:      l,
		connections: map[pkgid.ID]privConn{},
	}, nil
}

type privConn struct {
	conn   Connection
	waitCh chan struct{}
}

type eventServer struct {
	mu sync.RWMutex

	logger *zap.Logger

	connections map[pkgid.ID]privConn
}

func (c *eventServer) ServeConnection(id pkgid.ID, conn Connection) error {
	pc := privConn{
		conn:   conn,
		waitCh: make(chan struct{}, 1),
	}

	c.mu.Lock()
	c.connections[id] = pc
	// TODO: check if conn already exists
	c.mu.Unlock()

	<-pc.waitCh

	return nil
}

func (c *eventServer) SendEvent(ctx context.Context, msg Event) error {
	c.mu.RLock()
	conn, ok := c.connections[msg.RecipientID]
	c.mu.RUnlock()

	if !ok {
		// TODO: should be recognizable error
		return fmt.Errorf("user %s is not connected", msg.RecipientID)
	}

	return conn.conn.SendEvent(ctx, msg)
}
