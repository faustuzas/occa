package eventserver

import (
	"context"

	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/eventserver/generated/proto/eventserverpb"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

var _ eventserverpb.EventServerServer = (*GRPCServer)(nil)

func NewGRPCServer(l *zap.Logger, chatServer EventServer) *GRPCServer {
	return &GRPCServer{
		chatServer: chatServer,

		logger: l,
	}
}

type GRPCServer struct {
	eventserverpb.UnimplementedEventServerServer

	chatServer EventServer
	logger     *zap.Logger
}

func (s *GRPCServer) Connect(req *eventserverpb.ConnectRequest, server eventserverpb.EventServer_ConnectServer) error {
	id := pkgid.FromString(req.UserId)

	s.logger.Info("new gRPC based user connected", zap.Stringer("id", id))

	return s.chatServer.ServeConnection(id, NewGRPCConnection(server))
}

type grpcConnection struct {
	sender eventserverpb.EventServer_ConnectServer
}

func NewGRPCConnection(sender eventserverpb.EventServer_ConnectServer) Connection {
	return &grpcConnection{
		sender: sender,
	}
}

func (g grpcConnection) SendEvent(ctx context.Context, msg Event) error {
	return g.sender.Send(&eventserverpb.Event{
		Payload: &eventserverpb.Event_DirectMessage{
			DirectMessage: &eventserverpb.DirectMessage{Message: msg.Content},
		},
	})
}
