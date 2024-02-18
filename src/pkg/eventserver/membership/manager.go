package membership

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	pkgetcd "github.com/faustuzas/occa/src/pkg/etcd"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
)

const (
	refreshInfoDuration   = 15 * time.Second
	eventServersNamespace = "/event_servers/"
)

type Manager interface {
	JoinCluster(ctx context.Context, serverInfoFn func(context.Context) (ServerInfo, error)) (<-chan struct{}, error)
	LeaveCluster(ctx context.Context) error
}

type manager struct {
	lease   *pkgetcd.LeasedClient
	closeCh chan struct{}

	i pkginstrument.Instrumentation
}

func NewManager(i pkginstrument.Instrumentation, client *clientv3.Client) (Manager, error) {
	return &manager{
		lease:   pkgetcd.NewLeasedClient(i, client, 15*time.Second),
		closeCh: make(chan struct{}),
		i:       i,
	}, nil
}

type ServerInfo struct {
	ID          string `json:"id"`
	GRPCAddress string `json:"grpcAddress"`
	HTTPAddress string `json:"httpAddress"`

	// TODO: some kind of load parameter
}

func (i *ServerInfo) Marshall() ([]byte, error) {
	return json.Marshal(i)
}

func (i *ServerInfo) Unmarshall(data []byte) error {
	return json.Unmarshal(data, &i)
}

// JoinCluster joins the event server cluster and regularly updates the information stored there.
// Note: serverInfoFn should be thread-safe.
func (m *manager) JoinCluster(ctx context.Context, serverInfoFn func(context.Context) (ServerInfo, error)) (<-chan struct{}, error) {
	if err := m.lease.Start(ctx); err != nil {
		return nil, fmt.Errorf("acquiring lease: %w", err)
	}

	if err := m.refreshInfo(ctx, serverInfoFn); err != nil {
		return nil, fmt.Errorf("joining the cluster: %w", err)
	}

	go m.refreshInfoLoop(serverInfoFn)

	return m.lease.LeaseLostC(), nil
}

func (m *manager) refreshInfoLoop(serverInfoFn func(context.Context) (ServerInfo, error)) {
	ticker := time.NewTicker(refreshInfoDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-m.lease.LeaseLostC():
			m.i.Logger.Warn("closing membership update loop because of lost lease")
			return
		case <-m.closeCh:
			m.i.Logger.Info("closing membership update loop because of close request")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), refreshInfoDuration)
		if err := m.refreshInfo(ctx, serverInfoFn); err != nil {
			m.i.Logger.Warn("failed to refresh server info", zap.Error(err))
		}
		cancel()
	}
}

func (m *manager) refreshInfo(ctx context.Context, serverInfoFn func(context.Context) (ServerInfo, error)) error {
	info, err := serverInfoFn(ctx)
	if err != nil {
		return fmt.Errorf("getting server info: %w", err)
	}

	if info.ID == "" {
		return fmt.Errorf("empty server id provided")
	}

	infoBytes, err := info.Marshall()
	if err != nil {
		return fmt.Errorf("marshaling server info: %w", err)
	}

	// TODO: need to have some kind of epoch counter for safe overrides
	if err = m.lease.Put(ctx, eventServersNamespace+info.ID, infoBytes); err != nil {
		return fmt.Errorf("putting key: %w", err)
	}

	return nil
}

func (m *manager) LeaveCluster(ctx context.Context) error {
	close(m.closeCh)

	return m.lease.Close(ctx)
}
