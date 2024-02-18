package client

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/faustuzas/occa/src/pkg/eventserver/membership"
	"github.com/faustuzas/occa/src/pkg/generated/proto/rteventspb"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
	pkgio "github.com/faustuzas/occa/src/pkg/io"
)

type Pool interface {
	pkgio.Closer

	ClientForServer(ctx context.Context, serverID string) (Client, error)
}

type pool struct {
	serverInfoResolver membership.ServerInfoResolver
	i                  pkginstrument.Instrumentation
}

func NewPool(i pkginstrument.Instrumentation, serverInfoResolver membership.ServerInfoResolver) Pool {
	return &pool{
		serverInfoResolver: serverInfoResolver,
		i:                  i,
	}
}

func (p *pool) ClientForServer(ctx context.Context, serverID string) (Client, error) {
	info, err := p.serverInfoResolver.Resolve(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("resolving server info: %w", err)
	}

	return newHTTPClient(info.HTTPAddress), nil
}

func (p *pool) Close(_ context.Context) error {
	return nil // noop for now
}

type Client interface {
	Send(ctx context.Context, event *rteventspb.Event) error
}

type httpClient struct {
	c            *pkghttp.Client
	pbMarshaller jsonpb.Marshaler
}

func newHTTPClient(address string) *httpClient {
	return &httpClient{
		c:            pkghttp.NewClient(address),
		pbMarshaller: jsonpb.Marshaler{},
	}
}

func (h *httpClient) Send(ctx context.Context, event *rteventspb.Event) error {
	var buff bytes.Buffer
	if err := h.pbMarshaller.Marshal(&buff, event); err != nil {
		return fmt.Errorf("marshalling event: %w", err)
	}

	resp, err := h.c.Post(ctx, "/send-event", buff.Bytes())
	if err != nil {
		return fmt.Errorf("sending HTTP request: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("received non-200 status code: %v", resp.StatusCode)
	}

	return nil
}
