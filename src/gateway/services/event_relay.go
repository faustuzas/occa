package services

import (
	"context"
	"fmt"

	esclient "github.com/faustuzas/occa/src/pkg/eventserver/client"
	"github.com/faustuzas/occa/src/pkg/eventserver/rtconn"
	"github.com/faustuzas/occa/src/pkg/generated/proto/rteventspb"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
)

type RealTimeEventRelay interface {
	Forward(ctx context.Context, recipientID pkgid.ID, event *rteventspb.Event) error
}

type realTimeEventRelay struct {
	i pkginstrument.Instrumentation

	serverResolver rtconn.ServerResolver
	esPool         esclient.Pool
}

func NewRealTimeEventRelay(i pkginstrument.Instrumentation, serverResolver rtconn.ServerResolver, esPool esclient.Pool) RealTimeEventRelay {
	return &realTimeEventRelay{
		i:              i,
		serverResolver: serverResolver,
		esPool:         esPool,
	}
}

func (r *realTimeEventRelay) Forward(ctx context.Context, recipientID pkgid.ID, event *rteventspb.Event) error {
	serverInfo, err := r.serverResolver.Resolve(ctx, recipientID)
	if err != nil {
		return fmt.Errorf("resolving user server: %w", err)
	}

	client, err := r.esPool.ClientForServer(ctx, serverInfo.ServerID)
	if err != nil {
		return fmt.Errorf("resolving client for server: %w", err)
	}

	if err = client.Send(ctx, event); err != nil {
		return fmt.Errorf("sending event: %w", err)
	}

	return nil
}
