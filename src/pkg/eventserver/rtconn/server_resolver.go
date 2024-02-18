package rtconn

import (
	"context"
	"encoding/json"
	"fmt"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
	pkgmemstore "github.com/faustuzas/occa/src/pkg/memstore"
)

type ServerInformation struct {
	ServerID string
}

// ServerResolver resolves which event server user is connected to.
type ServerResolver interface {
	Resolve(ctx context.Context, userID pkgid.ID) (ServerInformation, error)
}

type serverResolver struct {
	store pkgmemstore.Store
	i     pkginstrument.Instrumentation
}

func NewServerResolver(i pkginstrument.Instrumentation, store pkgmemstore.Store) ServerResolver {
	return &serverResolver{
		store: store,
		i:     i,
	}
}

func (s *serverResolver) Resolve(ctx context.Context, userID pkgid.ID) (ServerInformation, error) {
	data, err := s.store.GetCollectionItem(ctx, connectionsNamespace, userID.String())
	if err != nil {
		return ServerInformation{}, fmt.Errorf("getting user connection info: %w", err)
	}

	var info ConnectionInfo
	if err = info.Unmarshall(data); err != nil {
		return ServerInformation{}, fmt.Errorf("unmarshalling conn info: %w", err)
	}

	return ServerInformation{
		ServerID: info.ServerID,
	}, nil
}

type ConnectionInfo struct {
	ServerID string `json:"serverID"`
}

func (i *ConnectionInfo) Marshall() ([]byte, error) {
	return json.Marshal(*i)
}

func (i *ConnectionInfo) Unmarshall(data []byte) error {
	return json.Unmarshal(data, i)
}
