package membership

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
)

type SelectedServer struct {
	ID          string
	GRPCAddress string
	HTTPAddress string
}

type ServerInfoResolver interface {
	Resolve(ctx context.Context, serverID string) (ServerInfo, error)
}

type ServerSelector interface {
	SelectServerForConnection(ctx context.Context) (ServerInfo, error)
}

// ServerRegistry is a registry of all currently available event servers.
// It should be used by event server clients to get information about them.
type ServerRegistry struct {
	mu sync.RWMutex

	etcdClient *clientv3.Client
	servers    map[string]ServerInfo

	closeCh chan struct{}

	i pkginstrument.Instrumentation
}

func NewServerRegistry(i pkginstrument.Instrumentation, etcdClient *clientv3.Client) *ServerRegistry {
	return &ServerRegistry{
		etcdClient: etcdClient,
		servers:    map[string]ServerInfo{},

		closeCh: make(chan struct{}),
		i:       i,
	}
}

func (r *ServerRegistry) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	getResp, err := r.etcdClient.Get(ctx, eventServersNamespace, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("getting initial servers: %w", err)
	}

	for _, kv := range getResp.Kvs {
		var info ServerInfo
		if err = info.Unmarshall(kv.Value); err != nil {
			return fmt.Errorf("unmarshaling server info: %w", err)
		}

		r.servers[info.ID] = info
	}

	r.i.Logger.Info("discovered event servers", zap.Any("servers", r.servers))

	watchStartRevision := getResp.Header.Revision + 1
	watchChan := r.etcdClient.Watch(ctx, eventServersNamespace, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
	go r.watchForUpdates(watchChan)

	return nil
}

func (r *ServerRegistry) Resolve(_ context.Context, serverID string) (ServerInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.servers[serverID]
	if !ok {
		return ServerInfo{}, fmt.Errorf("server not found in the registry")
	}

	return info, nil
}

func (r *ServerRegistry) SelectServerForConnection(_ context.Context) (ServerInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.servers) == 0 {
		return ServerInfo{}, fmt.Errorf("no servers in the registry")
	}

	// TODO: find a better way to select a server based on load, and more efficient of course
	var (
		num = rand.Intn(len(r.servers))
		i   int
	)
	for _, s := range r.servers {
		if i == num {
			return s, nil
		}
		i++
	}
	panic("not reachable")
}

// TODO: current implementation is not resistant of transient etcd failures, add retries
func (r *ServerRegistry) watchForUpdates(ch clientv3.WatchChan) {
	for {
		select {
		case <-r.closeCh:
			r.i.Logger.Info("closing server registry")
			return
		case watchEvent := <-ch:
			if watchEvent.Canceled {
				r.i.Logger.Info("watch got cancelled")
				return
			}

			if err := watchEvent.Err(); err != nil {
				r.i.Logger.Error("watch received an error", zap.Error(err))
				return
			}

			for _, e := range watchEvent.Events {
				switch e.Type {
				case clientv3.EventTypePut:
					var info ServerInfo
					if err := info.Unmarshall(e.Kv.Value); err != nil {
						r.i.Logger.Error("failed to unmarshal server info from watcher", zap.Error(err))
						continue
					}

					r.mu.Lock()
					_, ok := r.servers[info.ID]
					r.servers[info.ID] = info
					r.mu.Unlock()

					if ok {
						r.i.Logger.Info("event server info updated", zap.String("serverId", info.ID))
					} else {
						r.i.Logger.Info("new event server added", zap.String("serverId", info.ID))
					}
				case clientv3.EventTypeDelete:
					serverID := string(e.Kv.Key[len(eventServersNamespace):])

					r.mu.Lock()
					if _, ok := r.servers[serverID]; !ok {
						r.mu.Unlock()
						r.i.Logger.Warn("trying to remove server which was not present", zap.String("serverId", serverID))
						continue
					}

					delete(r.servers, serverID)
					r.mu.Unlock()

					r.i.Logger.Info("removed server", zap.String("serverId", serverID))
				}
			}
		}
	}
}

func (r *ServerRegistry) Close(_ context.Context) error {
	close(r.closeCh)

	return nil
}
