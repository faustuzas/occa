package rtconn

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	multierr "github.com/hashicorp/go-multierror"
	"go.uber.org/zap"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
	pkgio "github.com/faustuzas/occa/src/pkg/io"
	"github.com/faustuzas/occa/src/pkg/memstore"
)

const (
	connectionsNamespace = "user-connections"

	heartBeatTTL      = 20 * time.Second
	heartBeatInterval = 10 * time.Second
)

type HeartBeater interface {
	pkgio.Closer

	LaunchForUser(userID pkgid.ID) error
	StopForUser(userId pkgid.ID) error
}

type heart struct {
	stopCh chan struct{}
}

type heartBeater struct {
	mu sync.Mutex

	serverID string
	hearts   map[pkgid.ID]heart

	memstore memstore.Store
	i        pkginstrument.Instrumentation
}

func NewHeartBeater(i pkginstrument.Instrumentation, serverID string, store memstore.Store) HeartBeater {
	return &heartBeater{
		serverID: serverID,

		memstore: store,
		i:        i,
	}
}

func (b *heartBeater) LaunchForUser(userID pkgid.ID) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.hearts[userID]; ok {
		return fmt.Errorf("user already has heartbeat")
	}

	info := ConnectionInfo{ServerID: b.serverID}
	data, err := info.Marshall()
	if err != nil {
		return fmt.Errorf("failed to marshall connection info: %w", err)
	}

	h := heart{stopCh: make(chan struct{})}
	b.hearts[userID] = h

	go b.heartBeat(userID.String(), data, h.stopCh)

	return nil
}

func (b *heartBeater) heartBeat(userID string, data []byte, stopCh chan struct{}) {
	ticker := time.NewTicker(addJitter(heartBeatInterval))
	defer ticker.Stop()

	ctx := context.Background()
	for {
		if err := b.memstore.SetCollectionItemWithTTL(ctx, connectionsNamespace, userID, data, heartBeatTTL); err != nil {
			b.i.Logger.Error("failed to store user connection information", zap.Error(err))
			return
		}

		select {
		case <-ticker.C:
		case <-stopCh:
			return
		}
	}
}

func (b *heartBeater) StopForUser(userId pkgid.ID) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.stopForUserWithLock(userId)
}

func (b *heartBeater) stopForUserWithLock(userID pkgid.ID) error {
	h, ok := b.hearts[userID]
	if !ok {
		return fmt.Errorf("user does not have heartbeat")
	}

	close(h.stopCh)
	delete(b.hearts, userID)

	return nil
}

func (b *heartBeater) Close(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var multiErr error
	for userID := range b.hearts {
		if err := b.stopForUserWithLock(userID); err != nil {
			multiErr = multierr.Append(multiErr, err)
		}
	}
	return multiErr
}

func addJitter(duration time.Duration) time.Duration {
	return duration + time.Duration(rand.Intn(int(duration)/10))
}
