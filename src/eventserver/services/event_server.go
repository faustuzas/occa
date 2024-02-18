package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	multierr "github.com/hashicorp/go-multierror"

	"github.com/faustuzas/occa/src/pkg/eventserver/rtconn"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
)

type Event struct {
	SenderID    pkgid.ID  `json:"senderID"`
	RecipientID pkgid.ID  `json:"recipientID"`
	Content     string    `json:"content"`
	SentAt      time.Time `json:"sentAt"`
}

type Connection interface {
	SendEvent(ctx context.Context, msg Event) error
}

type EventServer interface {
	SendEvent(ctx context.Context, msg Event) error
	ServeConnection(id pkgid.ID, connection Connection) error
	InitiateShutdown(ctx context.Context) error
}

func NewEventServer(i pkginstrument.Instrumentation, heartBeater rtconn.HeartBeater) (EventServer, error) {
	return &eventServer{
		connections: map[pkgid.ID]privConn{},
		heartBeater: heartBeater,

		i: i,
	}, nil
}

type privConn struct {
	conn   Connection
	waitCh chan struct{}
}

type eventServer struct {
	mu          sync.RWMutex
	connections map[pkgid.ID]privConn

	heartBeater rtconn.HeartBeater

	i pkginstrument.Instrumentation
}

func (s *eventServer) ServeConnection(userID pkgid.ID, conn Connection) error {
	pc := privConn{
		conn:   conn,
		waitCh: make(chan struct{}, 1),
	}

	s.mu.Lock()
	s.connections[userID] = pc
	// TODO: check if conn already exists
	s.mu.Unlock()

	if err := s.heartBeater.LaunchForUser(userID); err != nil {
		return fmt.Errorf("failed to launch heart beater")
	}

	// TODO: find where connection terminates
	<-pc.waitCh

	return nil
}

func (s *eventServer) SendEvent(ctx context.Context, msg Event) error {
	s.mu.RLock()
	conn, ok := s.connections[msg.RecipientID]
	s.mu.RUnlock()

	if !ok {
		// TODO: should be recognizable error
		return fmt.Errorf("user %s is not connected", msg.RecipientID)
	}

	return conn.conn.SendEvent(ctx, msg)
}

func (s *eventServer) InitiateShutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	errCh := make(chan error, len(s.connections))
	for _, conn := range s.connections {
		go func(conn privConn) {
			// TODO: sent reconnect event
			//errCh <- conn.conn.SendEvent(ctx, )

			errCh <- nil
			close(conn.waitCh)
		}(conn)
	}

	var err error
	for i := 0; i < len(s.connections); i++ {
		select {
		case e := <-errCh:
			if e != nil {
				err = multierr.Append(err, e)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}
